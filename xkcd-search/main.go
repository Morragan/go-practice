package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

const xkcdURL = "https://xkcd.com/%d/info.0.json"

type Comic struct {
	Month      string `json:"month"`
	Num        int    `json:"num"`
	Link       string `json:"link"`
	Year       string `json:"year"`
	News       string `json:"news"`
	SafeTitle  string `json:"safe_title"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
	Title      string `json:"title"`
	Day        string `json:"day"`
}

func getExistingComics(outputFile string) []Comic {
	var comics []Comic

	outFile, err := os.OpenFile(outputFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", outputFile, err)
		return nil
	}
	defer outFile.Close()

	contents, err := io.ReadAll(outFile)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", outputFile, err)
		return nil
	}

	if len(contents) == 0 {
		fmt.Println("File is empty, initializing with []")
		comics = []Comic{}
	} else {
		err = json.Unmarshal(contents, &comics)
		if err != nil {
			fmt.Printf("Error unmarshalling JSON from file %s: %v\n", outputFile, err)
			return nil
		}
	}

	return comics
}

func saveComicsToFile(comics []Comic, outputFile string) {
	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error opening file %s for writing: %v\n", outputFile, err)
		return
	}
	defer file.Close()

	data, err := json.MarshalIndent(comics, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling comics to JSON: %v\n", err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		fmt.Printf("Error writing comics to file %s: %v\n", outputFile, err)
		return
	}
}

func downloadComics(outputFile string) []Comic {
	comics := getExistingComics(outputFile)
	if comics == nil {
		comics = []Comic{}
	}

	// count occurences of 404 errors; stop after 2 consecutive ones, because the comic 404 returns error 404
	for comicId, notFoundOccurrences := 1, 0; notFoundOccurrences < 2; comicId++ {
		existingComicIndex := slices.IndexFunc(comics, func(c Comic) bool {
			return c.Num == comicId
		})

		if existingComicIndex != -1 {
			notFoundOccurrences = 0
			fmt.Printf("Comic %d already exists, skipping...\n", comicId)
			continue
		}

		comicURL := fmt.Sprintf(xkcdURL, comicId)
		resp, err := http.Get(comicURL)

		if err != nil {
			fmt.Printf("Error fetching comic %d: %v. Stopping\n", comicId, err)
			break
		}

		if resp.StatusCode == http.StatusNotFound {
			notFoundOccurrences++
			fmt.Printf("Comic %d not found\n", comicId)
			continue
		}

		notFoundOccurrences = 0
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Comic %d returned status code %d\n", comicId, resp.StatusCode)
			continue
		}

		var comic Comic
		err = json.NewDecoder(resp.Body).Decode(&comic)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Error decoding JSON for comic %d: %v\n", comicId, err)
			continue
		}
		fmt.Printf("Comic %d downloaded: %s\n", comic.Num, comic.Title)
		comics = append(comics, comic)
	}

	saveComicsToFile(comics, outputFile)

	return comics
}

func findComicsBySearchTerm(term string, comics []Comic) []Comic {
	foundComics := []Comic{}
	for _, comic := range comics {
		if strings.Contains(strings.ToLower(comic.Transcript), strings.ToLower(term)) ||
			strings.Contains(strings.ToLower(comic.Title), strings.ToLower(term)) ||
			strings.Contains(strings.ToLower(comic.SafeTitle), strings.ToLower(term)) ||
			strings.Contains(strings.ToLower(comic.Alt), strings.ToLower(term)) {

			foundComics = append(foundComics, comic)
		}
	}

	return foundComics
}

func main() {
	isDownloadMode := flag.Bool("d", false, "If true, download the comics")
	outputFile := flag.String("o", "xkcd.json", "File to save the comics to")

	searchTerm := flag.String("s", "", "Search term for comics")

	flag.Parse()

	if !*isDownloadMode && *searchTerm == "" {
		fmt.Println("Please provide a search term or use the download mode.")
		return
	}

	var comics []Comic
	if *isDownloadMode {
		comics = downloadComics(*outputFile)
	}

	if *searchTerm != "" {
		if comics == nil {
			comics = getExistingComics(*outputFile)
			if comics == nil {
				fmt.Println("Downloading comics failed. Check your internet connection and the xkcd API.")
				return
			}
		}

		comicsWithTerm := findComicsBySearchTerm(*searchTerm, comics)
		if len(comicsWithTerm) == 0 {
			fmt.Printf("No comics found with the term '%s'.\n", *searchTerm)
			return
		}

		fmt.Printf("Found %d comics with the term '%s':\n", len(comicsWithTerm), *searchTerm)
		for _, comic := range comicsWithTerm {
			fmt.Printf("%d. %s - %s\n", comic.Num, comic.Title, comic.Img)
		}
	}
}
