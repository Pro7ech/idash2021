package main

import (
	"log"
	"os"
)

// Cleans the temporary files, which includes the keys, encrypted data and results.
func main() {

	cleanFolder("keys/")
	cleanFolder("temps/")
	cleanFolder("results/")

	//if err := os.Remove("results/curve.png"); err != nil {
	//	log.Println(err)
	//}
}

func cleanFolder(folderPath string) {

	folder, _ := os.Open(folderPath)
	files, _ := folder.Readdir(0)

	for i := range files {

		if files[i].Name() != "donotremove.txt" {
			if err := os.Remove(folderPath + files[i].Name()); err != nil {
				log.Println(err)
			}
		}
	}
}
