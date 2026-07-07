package util

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// 大模型调用超时较长（歌词整段翻译可能较慢）
var llmHTTPClient = &http.Client{Timeout: 180 * time.Second}

// LLMEnabled 返回是否配置了可用的大模型凭据。
func LLMEnabled() bool {
	return OpenAIAPIKey != "" && OpenAIBaseURL != "" && OpenAIModel != ""
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// chatComplete 调用 OpenAI 兼容的 /chat/completions 接口，返回助手回复文本。
func chatComplete(system, user string) (string, error) {
	if !LLMEnabled() {
		return "", errors.New("LLM not configured")
	}
	reqBody := chatRequest{
		Model: OpenAIModel,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0,
		Stream:      false,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.WithStack(err)
	}
	endpoint := OpenAIBaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", errors.WithStack(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OpenAIAPIKey)

	resp, err := llmHTTPClient.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("LLM HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", errors.WithStack(err)
	}
	if parsed.Error != nil {
		return "", errors.New(parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

const translateSystemPrompt = `你是专业的歌词翻译。将用户提供的每一行歌词翻译成简体中文。
严格要求：
1. 输出的行数必须与输入完全一致，第 i 行输出对应第 i 行输入。
2. 不要添加行号、注释、空行、时间戳或任何额外文字。
3. 对已经是中文的行，原样输出该行中文。
4. 只输出翻译结果，不要输出任何解释。`

const romajiSystemPrompt = `你是专业的歌词罗马音转写。将用户提供的每一行日语或韩语歌词转写为罗马音（日语用平文式 Hepburn 罗马字，韩语用韩语罗马字）。
严格要求：
1. 输出的行数必须与输入完全一致，第 i 行输出对应第 i 行输入。
2. 不要添加行号、注释、空行、时间戳或任何额外文字。
3. 罗马音使用小写拉丁字母，词间用空格分隔。
4. 只输出罗马音结果，不要输出任何解释。`

// stripCodeFence 去除模型可能包裹的 Markdown 代码块围栏。
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// llmTransform 用给定系统提示逐行转换文本，返回与输入等长的结果切片。
// 若行数与输入不一致（模型幻觉/漏行/多行），返回 nil，由调用方放弃本次更新。
func llmTransform(system string, lines []string) ([]string, error) {
	content, err := chatComplete(system, strings.Join(lines, "\n"))
	if err != nil {
		return nil, err
	}
	content = stripCodeFence(content)
	out := strings.Split(content, "\n")
	// 去除首尾空行（输入均为非空行，模型偶尔会多出首尾空行）
	for len(out) > 0 && strings.TrimSpace(out[0]) == "" {
		out = out[1:]
	}
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	if len(out) != len(lines) {
		return nil, nil // 行数不匹配：验证失败，防止幻觉污染
	}
	return out, nil
}

// buildBlended 用原文行与「行索引→译文」映射重建混合歌词。
func buildBlended(original []LyricLine, trans map[int]string) string {
	var sb strings.Builder
	for i, ln := range original {
		sb.WriteString(ln.Timestamp)
		sb.WriteString(ln.Text)
		sb.WriteString("\n")
		if t, ok := trans[i]; ok && strings.TrimSpace(t) != "" {
			sb.WriteString(ln.Timestamp)
			sb.WriteString(t)
			sb.WriteString("\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// CompleteLyric 检查并补齐一条歌词的翻译与罗马音。
//   - 返回新的（可能已混合中文翻译的）主歌词、新的罗马音、是否发生变更、错误。
//   - 若歌词已完备或未配置模型，changed=false 且原样返回。
//   - 所有模型输出都会经过「行数一致」校验，校验失败则跳过该项更新（防止幻觉污染数据）。
func CompleteLyric(lyric, romaji string) (newLyric string, newRomaji string, changed bool, err error) {
	newLyric, newRomaji = lyric, romaji
	if !LLMEnabled() {
		return newLyric, newRomaji, false, nil
	}

	original, translation := originalAndTranslation(lyric)
	lang := detectLang(joinOriginalText(original))

	needTrans := (lang == LangOther || lang == LangJapanese || lang == LangKorean) && !hasChineseTranslation(translation)
	needRoma := (lang == LangJapanese || lang == LangKorean) && !hasRomaji(romaji)
	if !needTrans && !needRoma {
		return newLyric, newRomaji, false, nil
	}

	// 收集非空原文行（附带其在 original 中的索引）
	var idxs []int
	var texts []string
	for i, ln := range original {
		if strings.TrimSpace(ln.Text) != "" {
			idxs = append(idxs, i)
			texts = append(texts, ln.Text)
		}
	}
	if len(texts) == 0 {
		return newLyric, newRomaji, false, nil
	}

	if needTrans {
		out, e := llmTransform(translateSystemPrompt, texts)
		if e != nil {
			return newLyric, newRomaji, changed, e
		}
		if out != nil { // 校验通过
			transByIdx := make(map[int]string, len(idxs))
			for k, i := range idxs {
				transByIdx[i] = strings.TrimSpace(out[k])
			}
			newLyric = buildBlended(original, transByIdx)
			changed = true
		}
	}

	if needRoma {
		out, e := llmTransform(romajiSystemPrompt, texts)
		if e != nil {
			return newLyric, newRomaji, changed, e
		}
		if out != nil { // 校验通过
			var sb strings.Builder
			for k, i := range idxs {
				r := strings.TrimSpace(out[k])
				if r == "" {
					continue
				}
				sb.WriteString(original[i].Timestamp)
				sb.WriteString(r)
				sb.WriteString("\n")
			}
			newRomaji = strings.TrimRight(sb.String(), "\n")
			changed = true
		}
	}

	return newLyric, newRomaji, changed, nil
}
