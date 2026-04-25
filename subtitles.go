package main

import (
	"bufio"
	"os"
	"strings"
)

// Styles calibrés pour 1920x1080 (PlayRes du CR), basés sur le modèle fansub fourni.
// Trebuchet MS 66px, contour 3px, ombre 3px, marges 75px.
var modelStyles = map[string]string{
	// Dialogue principal
	"Default": "Style: Default,Trebuchet MS,66,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,3,3,2,75,75,75,1",
	// Dialogue en italique
	"Italique": "Style: Italique,Trebuchet MS,66,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,-1,0,0,100,100,0,0,1,3,3,2,75,75,75,1",
	// Deux personnages (alignement haut)
	"TiretsDefault": "Style: TiretsDefault,Trebuchet MS,66,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,3,3,1,75,75,75,1",
	// Deux personnages italique
	"TiretsItalique": "Style: TiretsItalique,Trebuchet MS,66,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,-1,0,0,100,100,0,0,1,3,3,1,75,75,75,1",
	// Panneaux/textes à l'écran — Arial, fond sombre, alignement libre
	"Sign": "Style: Sign,Arial,63,&H00FFFFFF,&H000000FF,&H00292929,&H00000000,-1,0,0,0,100,100,0,0,1,3,0,8,75,75,75,1",
}

// isSignStyle retourne true si le nom de style correspond à un panneau/texte à l'écran
func isSignStyle(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "sign") ||
		strings.Contains(lower, "titre") ||
		strings.Contains(lower, "title") ||
		strings.Contains(lower, "caption")
}

// styleSubtitles reécrit le fichier ASS avec les styles du modèle
func styleSubtitles(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	inStyles := false
	stylesWritten := map[string]bool{}

	for scanner.Scan() {
		line := scanner.Text()

		// Repère la section styles
		if strings.HasPrefix(line, "[V4+ Styles]") {
			inStyles = true
			lines = append(lines, line)
			continue
		}
		// Fin de la section styles
		if inStyles && strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[V4+ Styles]") {
			// Injecte les styles du modèle qui n'ont pas encore été écrits
			for name, style := range modelStyles {
				if !stylesWritten[name] {
					lines = append(lines, style)
					stylesWritten[name] = true
				}
			}
			inStyles = false
		}

		if inStyles && strings.HasPrefix(line, "Style:") {
			// Extrait le nom du style CR (ex: "Default", "Default - ja-JP", "Signs", ...)
			rest := strings.TrimPrefix(line, "Style:")
			parts := strings.SplitN(strings.TrimSpace(rest), ",", 2)
			crName := strings.TrimSpace(parts[0])

			// Détermine quel style du modèle appliquer
			var modelKey string
			switch {
			case isSignStyle(crName):
				modelKey = "Sign"
			case strings.Contains(strings.ToLower(crName), "italic"):
				modelKey = "Italique"
			default:
				modelKey = "Default"
			}

			if styled, ok := modelStyles[modelKey]; ok && !stylesWritten[crName] {
				// Garde le nom original du style CR mais applique nos paramètres
				params := strings.SplitN(styled, ",", 2)
				lines = append(lines, "Style: "+crName+","+params[1])
				stylesWritten[crName] = true
			}
			continue
		}

		// Corrige le ScriptInfo si nécessaire
		if strings.HasPrefix(line, "ScaledBorderAndShadow:") {
			lines = append(lines, "ScaledBorderAndShadow: yes")
			continue
		}
		if strings.HasPrefix(line, "WrapStyle:") {
			lines = append(lines, "WrapStyle: 0")
			continue
		}

		lines = append(lines, line)
	}
	f.Close()

	// Écriture du fichier corrigé
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		w.WriteString(l + "\n")
	}
	return w.Flush()
}
