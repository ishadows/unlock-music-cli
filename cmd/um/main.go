package main

import (
	"errors"
	"github.com/unlock-music/cli/algo/common"
	_ "github.com/unlock-music/cli/algo/kgm"
	_ "github.com/unlock-music/cli/algo/kwm"
	_ "github.com/unlock-music/cli/algo/ncm"
	_ "github.com/unlock-music/cli/algo/qmc"
	_ "github.com/unlock-music/cli/algo/tm"
	_ "github.com/unlock-music/cli/algo/xm"
	"github.com/unlock-music/cli/internal/logging"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var AppVersion = "0.0.3"

func main() {
	app := cli.App{
		Name:     "Unlock Music CLI",
		HelpName: "um",
		Usage:    "Unlock your encrypted music file https://github.com/unlock-music/cli",
		Version:  AppVersion,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "input", Aliases: []string{"i"}, Usage: "path to input file or dir", Required: true},
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "path to output dir", Required: true},
		},
		Action:          appMain,
		Copyright:       "Copyright (c) 2020 Unlock Music https://github.com/unlock-music/cli/blob/master/LICENSE",
		HideHelpCommand: true,
		UsageText:       "um -i /path/to/input -o /path/to/output/dir",
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func appMain(c *cli.Context) error {
	input := c.String("input")
	output := c.String("output")
	inputStat, err := os.Stat(input)
	if err != nil {
		return err
	}

	outputStat, err := os.Stat(output)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(output, 0755)
		}
		if err != nil {
			return err
		}
	} else if !outputStat.IsDir() {
		return errors.New("output should be a writable directory")
	}

	if inputStat.IsDir() {
		return dealDirectory(input, output)
	} else {
		ext := strings.TrimLeft(filepath.Ext(inputStat.Name()), ".")
		allDec := common.GetDecoder(ext)
		if len(allDec) == 0 {
			logging.Log().Fatal("skipping while no suitable decoder")
		}
		return tryDecFile(input, output, allDec)
	}

}
func dealDirectory(inputDir string, outputDir string) error {
	items, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		ext := strings.TrimLeft(filepath.Ext(item.Name()), ".")
		allDec := common.GetDecoder(ext)
		if len(allDec) == 0 {
			logging.Log().Info("skipping while no suitable decoder", zap.String("file", item.Name()))
			continue
		}

		err := tryDecFile(filepath.Join(inputDir, item.Name()), outputDir, allDec)
		if err != nil {
			logging.Log().Error("conversion failed", zap.String("source", item.Name()))
		}
	}
	return nil
}

func tryDecFile(inputFile string, outputDir string, allDec []common.NewDecoderFunc) error {
	file, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	var dec common.Decoder
	for _, decFunc := range allDec {
		dec = decFunc(file)
		if err := dec.Validate(); err == nil {
			break
		}
		logging.Log().Warn("try decode failed", zap.Error(err))
		dec = nil
	}
	if dec == nil {
		return errors.New("no any decoder can resolve the file")
	}
	if err := dec.Decode(); err != nil {
		return errors.New("failed while decoding: " + err.Error())
	}

	outData := dec.GetAudioData()
	outExt := dec.GetAudioExt()
	if outExt == "" {
		if ext, ok := common.SniffAll(outData); ok {
			outExt = ext
		} else {
			outExt = ".mp3"
		}
	}
	filenameOnly := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	outPath := filepath.Join(outputDir, filenameOnly+outExt)
	err = os.WriteFile(outPath, outData, 0644)
	if err != nil {
		return err
	}
	logging.Log().Info("successfully converted",
		zap.String("source", inputFile), zap.String("destination", outPath))
	return nil
}
