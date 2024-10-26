package main

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func replaceBadWords(body string) string {

	bodySlices := strings.Split(strings.ToLower(body), " ")

	for i, word := range bodySlices {
		switch word {
		case "kerfuffle":
			bodySlices[i] = "****"
		case "sharbert":
			bodySlices[i] = "****"
		case "fornax":
			bodySlices[i] = "****"
		case "i":
			bodySlices[i] = "I"
		case "chirpy.":
			bodySlices[i] = "Chirpy."
		case "mastodon":
			bodySlices[i] = "Mastodon"
		}

	}
	bodySlices[0] = cases.Title(language.English).String(bodySlices[0])
	cleanString := strings.Join(bodySlices, " ")

	return cleanString
}
