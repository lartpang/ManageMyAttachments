package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/akamensky/argparse"
)

// regexp并未支持Lookahead功能。
var EXT_PATTERN = regexp.MustCompile(`\.(jpg|png|jpeg|bmp|gif)$`)
var LINK_PATTERN = regexp.MustCompile(`!\[.*?\]\((.*?\.(jpg|png|jpeg|bmp|gif))\)`)

func readFileByLine(path string, container []string) ([]string, error) {
	fi, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return container, err
	}
	defer fi.Close()

	buffer_scanner := bufio.NewScanner(fi)
	for buffer_scanner.Scan() {
		line := buffer_scanner.Text()
		found_groups := LINK_PATTERN.FindAllStringSubmatch(line, -1)
		for _, found_group := range found_groups {
			link := found_group[1]
			if !strings.HasPrefix(link, "http") {
				dir_path := filepath.Dir(path)
				abs_file_path := filepath.Join(dir_path, link)
				container = append(container, abs_file_path)
			}
		}
	}
	return container, nil
}

func getPathsFromFileAndDir(
	target_dir string,
	target_folder string,
	paths_from_target_folder []string,
	paths_from_dir_files []string,
) ([]string, []string) {
	items, _ := os.ReadDir(target_dir)
	for _, item := range items {
		file_name := item.Name()
		file_path := filepath.Join(target_dir, file_name)

		if filepath.Base(target_dir) == target_folder && !item.IsDir() {
			// 已进到附件文件夹中
			if len(EXT_PATTERN.FindString(file_name)) > 0 {
				paths_from_target_folder = append(paths_from_target_folder, file_path)
			}
		} else {
			// 已进到其他文件夹中
			if item.IsDir() {
				paths_from_target_folder, paths_from_dir_files = getPathsFromFileAndDir(
					file_path,
					target_folder,
					paths_from_target_folder,
					paths_from_dir_files,
				)
			} else {
				if strings.HasSuffix(file_name, ".md") {
					paths_from_dir_files, _ = readFileByLine(file_path, paths_from_dir_files)
				}
			}
		}
	}
	return paths_from_target_folder, paths_from_dir_files
}

func removeDuplicatesOnStringSlice(src []string, verbose bool) []string {
	tgt := []string{}
	temp := map[string]struct{}{}

	for _, val := range src {
		if _, isInMap := temp[val]; !isInMap {
			temp[val] = struct{}{}
			tgt = append(tgt, val)
		}
	}

	if verbose {
		fmt.Printf("Length of Src: %d\nLength of Tgt: %d\n", len(src), len(tgt))
	}
	return tgt
}

func findDifferentStringItems(src0 []string, src1 []string, verbose bool) []string {
	tgt := []string{}
	temp := map[string]struct{}{}

	// 使用src0构建字典
	for _, val := range src0 {
		if _, isInMap := temp[val]; !isInMap {
			temp[val] = struct{}{}
		}
	}
	// 使用src1寻找差异项
	for _, val := range src1 {
		if _, isInMap := temp[val]; !isInMap {
			tgt = append(tgt, val)
		}
	}

	if verbose {
		fmt.Printf(
			"Length of Src0: %d\nLength of Src1: %d\nLength of DifferenceSet: %d\n",
			len(src0),
			len(src1),
			len(tgt),
		)
	}
	return tgt
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

func isEmpty(dir_path string) bool {
	dir, _ := os.ReadDir(dir_path)
	if len(dir) == 0 {
		return true
	} else {
		return false
	}
}

func main() {
	// 反引号包含内容不转义，而双引号包含是会转义的
	parser := argparse.NewParser("print", "Prints provided string to stdout")
	var target_folder *string = parser.String("", `target-folder`, &argparse.Options{Default: `assets`, Help: `The name of the attachment folder.`})
	var target_dir *string = parser.String("", `target-root`, &argparse.Options{Required: true, Help: `The root directory for storing your documents.`})
	var mode *string = parser.Selector("", "mode", []string{"update_location", "list_useless", "delete_useless"},
		&argparse.Options{Default: `list_useless`, Help: `
            - update_location: Move the image file to the image location indicated in the markdown document.
            - list_useless: List unreferenced image files.
            - delete_useless: Delete unreferenced image files.`})
	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		os.Exit(-1)
	}

	paths_from_target_folder := []string{}
	paths_from_dir_files := []string{}
	paths_from_target_folder, paths_from_dir_files = getPathsFromFileAndDir(
		*target_dir,
		*target_folder,
		paths_from_target_folder,
		paths_from_dir_files,
	)
	paths_from_target_folder = removeDuplicatesOnStringSlice(paths_from_target_folder, false)
	paths_from_dir_files = removeDuplicatesOnStringSlice(paths_from_dir_files, false)
	paths_difference_from_target_folder := findDifferentStringItems(
		paths_from_dir_files,
		paths_from_target_folder,
		true,
	)
	paths_difference_from_dir_files := findDifferentStringItems(
		paths_from_target_folder,
		paths_from_dir_files,
		true,
	)

	if *mode == `update_location` {
		// 将对应错误的文件中的路径对应的文件从附件目录中进行索引，并将其移动到正确的目录下
		for _, path_from_dir_files := range paths_difference_from_dir_files {
			base_name_from_file := filepath.Base(path_from_dir_files)
			dir_path_from_file := filepath.Dir(path_from_dir_files)
			for _, path_from_folder := range paths_from_target_folder {
				base_name_from_dir := filepath.Base(path_from_folder)
				dir_path_from_dir := filepath.Dir(path_from_folder)
				if base_name_from_dir == base_name_from_file {
					fmt.Println("Incorrect Path", path_from_folder)
					if !isExist(dir_path_from_file) {
						// 递归创建文件夹
						err := os.MkdirAll(dir_path_from_file, os.ModePerm)
						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Println(dir_path_from_file, "does not exist, let's create it.")
						}
					}
					err := os.Rename(path_from_folder, path_from_dir_files)
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("MOVE ", path_from_folder, " TO ", path_from_dir_files)
					}

					// 移动后，如果文件夹变空，就删掉
					if isExist(dir_path_from_dir) && isEmpty(dir_path_from_dir) {
						err := os.Remove(dir_path_from_dir)
						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Println(dir_path_from_dir, "is empty, delete it.")
						}
					}

					break
				}
			}

		}
	} else if *mode == `list_useless` {
		fmt.Printf("paths_useless:\n")
		for idx, path := range paths_difference_from_target_folder {
			fmt.Println(idx, path)
		}
		fmt.Printf("paths_from_target_folder %d", len(paths_from_target_folder))
		fmt.Printf("paths_from_dir_files %d", len(paths_from_dir_files))
	} else if *mode == `delete_useless` {
		for _, path_from_folder := range paths_difference_from_target_folder {
			dir_path_from_folder := filepath.Dir(path_from_folder)
			err := os.Remove(path_from_folder)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(path_from_folder, "is unreferenced, delete it.")
			}
			// 移动后，如果文件夹变空，就删掉
			if isExist(dir_path_from_folder) && isEmpty(dir_path_from_folder) {
				err := os.Remove(dir_path_from_folder)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(dir_path_from_folder, "is empty, delete it.")
				}
			}
		}
	} else {
		fmt.Println("NotImplementedError")
	}
}
