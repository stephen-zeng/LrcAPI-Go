package request

import (
	"lrcAPI/file"
	"lrcAPI/processor"
)

type Request struct {
	Processor processor.Processor
	File      file.File
}
