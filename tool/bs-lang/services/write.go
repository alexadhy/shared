package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/pkg/errors"
	"github.com/tidwall/pretty"
	"go.amplifyedge.org/shared-v2/tool/bs-lang/services/config"
	"go.amplifyedge.org/shared-v2/tool/bs-lang/utils"
)

// WriteDataDumpFiles exported
func WriteDataDumpFiles(csvFilePath string, jsonDirPath string, sheet string) error {
	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		log.Println(sheet, " : ", "Cannot open file:"+csvFilePath, err)
		return errors.Wrap(err, "Cannot open file:"+csvFilePath)

	}
	// get csf file content
	csvFileContent, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return errors.New("Cannot read file:" + csvFilePath)
	}
	keys := csvFileContent[0][0:]
	// walk content for each lang
	var data []map[string]string
	for rowIndex, row := range csvFileContent[1:][0:] {
		data = append(data, map[string]string{})
		for columnIndex, key := range keys {
			if strings.Contains(key, "_url") && strings.TrimSpace(row[columnIndex]) != "" {
				if strings.Contains(row[columnIndex], ",") {
					var paths []string
					urls := strings.Split(row[columnIndex], ",")
					for _, url := range urls {
						paths = append(paths, utils.DownloadURL(strings.TrimSpace(url), sheet))
					}
					data[rowIndex][key] = strings.Join(paths, ",")
				} else {
					data[rowIndex][key] = utils.DownloadURL(strings.TrimSpace(row[columnIndex]), sheet)
				}

			} else {
				data[rowIndex][key] = row[columnIndex]
			}

		}
	}
	FileName := sheet + ".json"

	FilePath := filepath.Join(jsonDirPath, FileName)

	AbsPath, err := filepath.Abs(FilePath)
	if err != nil {
		log.Println(sheet, " : ", "Cannot get path specified: \""+AbsPath+"\"", err)
		return errors.New("Cannot get path specified:" + FilePath)
	}

	utils.CreateFile(AbsPath)

	file, err := os.OpenFile(AbsPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(sheet, " : ", "Cannot open file: \""+AbsPath+"\"", err)
		return errors.Wrap(err, "Cannot open file: \""+AbsPath+"\"")
	}

	err = file.Truncate(0)
	if err != nil {
		return errors.New("Cannot truncate file:" + sheet)
	}
	encodedJSON, _ := json.Marshal(data)

	_, err = file.Write(prettyJSON(encodedJSON))
	if err != nil {
		return errors.New("Cannot write to file:" + sheet)
	}
	err = file.Close()
	if err != nil {
		return errors.New("Cannot Close to file:" + sheet)
	}
	return nil
}

func writeOutFile(data, outFilePath string) error {

	file, err := os.OpenFile(outFilePath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	// write to lang file
	err = file.Truncate(0)

	if err != nil {
		return fmt.Errorf("Cannot truncate file: %v", err)
	}

	_, err = file.Write([]byte(data))

	if err != nil {
		return fmt.Errorf("Cannot write to file: %v", err)
	}

	return file.Close()
}

// WriteLanguageFiles exported
func WriteLanguageFiles(csvFilePath string, jsonDirPath string, sheet string) error {
	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		log.Println(sheet, " : ", "Cannot open file:"+csvFilePath, err)
		return errors.Wrap(err, sheet+" : "+"Cannot open file:"+csvFilePath)

	}
	// get csf file content
	csvFileContent, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return errors.New("Cannot read file:" + csvFilePath)
	}

	langFileName := "labels.json"

	langFilePath := filepath.Join(jsonDirPath, langFileName)

	langAbsPath, err := filepath.Abs(langFilePath)
	if err != nil {
		log.Println(sheet, " : ", "Cannot get path specified: \""+langAbsPath+"\"", err)
		return errors.New("Cannot get path specified:" + langFilePath)
	}

	utils.CreateFile(langAbsPath)

	file, err := os.OpenFile(langAbsPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(sheet, " : ", "Cannot open file: \""+langAbsPath+"\"", err)
		return errors.Wrap(err, sheet+" : "+"Cannot open file: \""+langAbsPath+"\"")
	}

	// log.Println("langAbsPath: \"" + langAbsPath + "\"")
	// os.Exit(1)

	err = file.Truncate(0)
	if err != nil {
		return err
	}
	mapFull := map[string]map[string]string{}
	// walk content for each lang
	for i, lang := range csvFileContent[0][1:] {

		mapLn := map[string]string{}
		log.Println("Language : ", lang, i)
		for j, row := range csvFileContent[1:] {
			// fmt.Println(csvFileContent[j+1][0], row[i+1])
			mapLn[csvFileContent[j+1][0]] = row[i+1]
		}
		mapFull[lang] = mapLn
	}
	encodedJSON, _ := json.Marshal(mapFull)
	// log.Println(string(encodedJSON))

	_, err = file.Write(prettyJSON(encodedJSON))
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

// WriteFiles Write files
func WriteFiles(csvFilePath string, config config.Config, cleanTagsDir, cleanTagsFileName string) error {

	// open csv file
	csvFileContent, err := utils.OpenCSVFile(csvFilePath)

	if err != nil {
		return err
	}

	err = utils.MkDirIfNotExists(config.OutDir)

	if err != nil {
		return fmt.Errorf("Cannot create dir: %v", err)
	}

	log.Println("Start generating data...")
	if config.Merge == "row" {
		return mergeRow(csvFileContent, config, cleanTagsDir, cleanTagsFileName)
	} else if config.Merge == "column" {
		return mergeColumns(csvFileContent, config, cleanTagsDir, cleanTagsFileName)
	} else if config.Merge == "cell" {
		return mergeCell(csvFileContent, config, cleanTagsDir, cleanTagsFileName)
	}

	return errors.New("Merge should be column or row")
}

// FormatJSON formats the JSON to be tidy and readable
func prettyJSON(json []byte) (result []byte) {
	result = pretty.Pretty(json)
	return result
}

// GenerateMultiLanguagesArbFilesFromJSONFiles generate arb files from json files
func GenerateMultiLanguagesArbFilesFromJSONFiles(dir, prefix, extFile, outExtFile string, full bool) error {

	if extFile == outExtFile {
		return errors.New("extension file and out file extension should not be the same")
	}

	fileInfos, err := ioutil.ReadDir(dir)

	if err != nil {
		return err
	}
	for _, file := range fileInfos {
		if !file.IsDir() {
			name, ext := getFileNameAndExtension(file.Name())

			if ext == extFile && strings.HasPrefix(name, prefix) {

				outDir := getPath(dir, name, outExtFile)
				p := getPath(dir, name, ext)
				data, err := ioutil.ReadFile(p)

				if err != nil {
					return err
				}

				if full {

					m, err := JSONMap(data)
					if err != nil {
						return err
					}
					data, err = m.ToJSON()
					if err != nil {
						return err
					}

					data = prettyJSON(data)

					err = writeOutFile(string(data), outDir)
					if err != nil {
						return err
					}
				} else {

					f, err := os.Create(outDir)
					if err != nil {
						return err
					}
					_, err = f.Write(data)
					if err != nil {
						return err
					}
					err = f.Close()
					if err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
}

// mlOutputConfig is the output configuration for GenerateMultiLanguageFilesFromTemplate
// type mlOutputConfig struct {
// 	templatePath string // template arb filepath
// 	outPath      string // output path
// 	ext          string // extension name
// 	separator string // separator for template words
// 	languages []string // languages to generate
// 	full bool //
// }

// GenerateMultiLanguageFilesFromTemplate write multilanguage json files
func GenerateMultiLanguageFilesFromTemplate(templatePath, outPath, fileName, ext string, languages []string, full bool, cachePath string) error {

	_, err := os.Stat(cachePath)
	if err != nil {
		defContent := map[string]interface{}{}
		data, err := json.Marshal(&defContent)
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(cachePath, data, 0755); err != nil {
			return err
		}
	}

	data, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}

	m := linkedhashmap.New()
	err = m.FromJSON(data)
	if err != nil {
		return err
	}

	wordsTranslated, err := getTemplateWords(m, config.TranslateTimeout, 3, languages, cachePath)
	if err != nil {
		return err
	}

	translatedMaps, err := getTranslatedMaps(wordsTranslated, m, full)

	if err != nil {
		return err
	}

	return writeOutFiles(translatedMaps, outPath, fileName, ext)
}

func writeOutFiles(translatedMaps *TranslatedMaps, outPath, fileName, ext string) error {

	for i, trMap := range translatedMaps.Maps {

		data, err := trMap.ToJSON()
		if err != nil {
			return err
		}

		name := fmt.Sprintf("%s_%s", fileName, translatedMaps.Langs[i])
		outPath := getPath(outPath, name, ext)
		data = prettyJSON(data)

		err = writeOutFile(string(data), outPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPath(dirPath, fileName, ext string) string {
	return path.Join(dirPath, fileName+"."+ext)
}

func getFileNameAndExtension(fileName string) (name string, ext string) {
	// extension
	origExt := filepath.Ext(fileName)
	if origExt != "" {
		ext = origExt[1:] // removes the '.'
	}
	// get the filename
	name = strings.TrimSuffix(fileName, ext)
	return
}
