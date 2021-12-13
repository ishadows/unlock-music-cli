package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/unlock-music/cli/algo/common"
	_ "github.com/unlock-music/cli/algo/kgm"
	_ "github.com/unlock-music/cli/algo/kwm"
	_ "github.com/unlock-music/cli/algo/ncm"
	_ "github.com/unlock-music/cli/algo/qmc"
	_ "github.com/unlock-music/cli/algo/tm"
	_ "github.com/unlock-music/cli/algo/xm"
	"github.com/unlock-music/cli/internal/logging"
)

var AppVersion = "v0.0.6"

func main() {
	app := cli.App{
		Name:     "Unlock Music CLI",
		HelpName: "um",
		Usage:    "Unlock your encrypted music file https://github.com/unlock-music/cli",
		Version:  fmt.Sprintf("%s (%s,%s/%s)", AppVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH),
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "input", Aliases: []string{"i"}, Usage: "path to input file or dir", Required: false},
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "path to output dir", Required: false},
			&cli.BoolFlag{Name: "skip-noop", Aliases: []string{"n"}, Usage: "skip noop decoder", Required: false, Value: true},
			&cli.BoolFlag{Name: "supported-ext", Usage: "Show supported file extensions and exit", Required: false, Value: false},
		},

		Action:          appMain,
		Copyright:       "Copyright (c) 2020 - 2021 Unlock Music https://github.com/unlock-music/cli/blob/master/LICENSE",
		HideHelpCommand: true,
		UsageText:       "um [-o /path/to/output/dir] [--extra-flags] [-i] /path/to/input",
	}
	err := app.Run(os.Args)
	if err != nil {
		logging.Log().Fatal("run app failed", zap.Error(err))
	}
}
func printSupportedExtensions() {
	exts := []string{}
	for ext := range common.DecoderRegistry {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	for _, ext := range exts {
		fmt.Printf("%s: %d\n", ext, len(common.DecoderRegistry[ext]))
	}
}
func appMain(c *cli.Context) (err error) {
	if c.Bool("supported-ext") {
		printSupportedExtensions()
		return nil
	}
	input := c.String("input")
	if input == "" {
		switch c.Args().Len() {
		case 0:
			input, err = os.Getwd()
			if err != nil {
				return err
			}
		case 1:
			input = c.Args().Get(0)
		default:
			return errors.New("please specify input file (or directory)")
		}
	}

	output := c.String("output")
	if output == "" {
		var err error
		output, err = os.Getwd()
		if err != nil {
			return err
		}
		if input == output {
			return errors.New("input and output path are same")
		}
	}

	skipNoop := c.Bool("skip-noop")

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
		return dealDirectory(input, output, skipNoop)
	} else {
		allDec := common.GetDecoder(inputStat.Name(), skipNoop)
		if len(allDec) == 0 {
			logging.Log().Fatal("skipping while no suitable decoder")
		}
		return tryDecFile(input, output, allDec)
	}

}
func dealDirectory(inputDir string, outputDir string, skipNoop bool) error {
	items, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		allDec := common.GetDecoder(item.Name(), skipNoop)
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
