# crunchyroll-downloader

Télécharge des animes depuis Crunchyroll et les exporte en fichiers `.mkv`.

Développé par [CuteTenshii](https://github.com/CuteTenshii/crunchyroll-downloader) — amélioré par **Pascool**.

---

## Prérequis

- [FFmpeg](https://www.ffmpeg.org/download.html) — doit être accessible dans le PATH ou dans le même dossier
- Un fichier `.wvd` **ou** les fichiers `client_id.bin` + `private_key.pem` (Widevine CDM)
- Un compte Crunchyroll (Premium requis pour le contenu Premium)

---

## Installation

1. Télécharge le binaire correspondant à ton OS depuis les [releases](https://github.com/CuteTenshii/crunchyroll-downloader/releases/latest)
2. Place `ffmpeg.exe` (Windows) ou `ffmpeg` (Linux/Mac) dans le même dossier, ou ajoute-le au PATH système
3. Place ton fichier `.wvd` (ou `client_id.bin` + `private_key.pem`) dans le même dossier
4. Récupère ton cookie `etp_rt` (voir section **Authentification** ci-dessous)

---

## Authentification

Le cookie `etp_rt` est nécessaire pour s'authentifier à Crunchyroll.

**Comment le récupérer :**
1. Va sur [crunchyroll.com](https://www.crunchyroll.com) et connecte-toi
2. Ouvre les DevTools (`Ctrl+Shift+I`)
3. **Firefox** : onglet *Stockage* → *Cookies*
4. **Chrome/Edge** : onglet *Application* → *Cookies*
5. Sélectionne le domaine Crunchyroll et copie la valeur du cookie `etp_rt`

---

## Utilisation

```
Usage:
  crdl-windows.exe [options]

Options:
  -url string
        URL d'un épisode ou d'une série Crunchyroll
  -urls string
        Chemin vers un fichier texte contenant une URL par ligne (batch)
  -etp-rt string
        Valeur du cookie "etp_rt" de ton compte (obligatoire)
  -season int
        Numéro de saison à télécharger (uniquement pour les URLs /series/)
  -audio-lang string
        Langue audio (défaut: "ja-JP")
  -subs-lang string
        Langue des sous-titres (défaut: "en-US")
  -video-quality string
        Qualité vidéo (défaut: "1080p") — valeurs: 360p, 480p, 720p, 1080p
  -audio-quality string
        Qualité audio (défaut: "192k") — valeurs: 96k, 128k, 192k
  -tag string
        Tag de release ajouté au nom de fichier (défaut: "Pascool")
```

---

## Exemples

### Télécharger un épisode unique
```shell
crdl-windows.exe \
  --url https://www.crunchyroll.com/watch/GR2PEGZPR/episode-title \
  --etp-rt TON_COOKIE
```

### Télécharger une saison complète
```shell
crdl-windows.exe \
  --url https://www.crunchyroll.com/series/GJ0H7Q5ZJ/hells-paradise \
  --season 1 \
  --etp-rt TON_COOKIE
```

### Télécharger toutes les saisons
```shell
crdl-windows.exe \
  --url https://www.crunchyroll.com/series/GJ0H7Q5ZJ/hells-paradise \
  --etp-rt TON_COOKIE
```

### Télécharger avec audio français (VF) et sous-titres français
```shell
crdl-windows.exe \
  --url https://www.crunchyroll.com/series/GJ0H7Q5ZJ/hells-paradise \
  --season 1 \
  --audio-lang fr-FR \
  --subs-lang fr-FR \
  --etp-rt TON_COOKIE
```

### Batch download (plusieurs URLs)
Crée un fichier `liste.txt` avec une URL par ligne :
```
https://www.crunchyroll.com/series/GJ0H7Q5ZJ/hells-paradise
https://www.crunchyroll.com/series/GEXH3WEP2/demon-slayer
```
Puis lance :
```shell
crdl-windows.exe --urls liste.txt --etp-rt TON_COOKIE
```

---

## Format de sortie

Les fichiers sont nommés selon la convention suivante :

```
SeriesTitle.S01E01.CR.WEBDL.VOSTFR.1080p.x265-Pascool.mkv
```

| Élément | Description |
|---|---|
| `CR` | Source : Crunchyroll |
| `WEBDL` | Type : stream web |
| `VOSTFR` | Sous-titres français, audio japonais |
| `Multi` | Audio japonais + VF disponible |
| `VF` | Audio français |
| `1080p` | Résolution vidéo |
| `x265` | Codec vidéo (après encodage) |
| `Pascool` | Tag de release (modifiable via `--tag`) |

Les fichiers sont organisés dans un dossier portant le nom de la série.

---

## Langues disponibles

| Code | Langue |
|---|---|
| `ja-JP` | Japonais (défaut) |
| `fr-FR` | Français |
| `en-US` | Anglais |
| `de-DE` | Allemand |
| `es-ES` | Espagnol (Espagne) |
| `es-419` | Espagnol (Amérique Latine) |
| `pt-BR` | Portugais (Brésil) |
| `it-IT` | Italien |
| `ru-RU` | Russe |
| `ar-SA` | Arabe |
| `ko-KR` | Coréen |
| `zh-CN` | Chinois simplifié |

---

## Resume (reprise automatique)

Si un téléchargement est interrompu, relance simplement la même commande. Les segments déjà téléchargés sont mis en cache dans `.crdl_cache/` et ne seront pas re-téléchargés. Le cache est supprimé automatiquement une fois l'épisode terminé.

---

## Build depuis les sources

**Prérequis :** [Go 1.21+](https://go.dev/dl/)

```shell
git clone https://github.com/CuteTenshii/crunchyroll-downloader
cd crunchyroll-downloader
go build .
```

---

## Fichier Widevine (.wvd)

Crunchyroll chiffre ses vidéos avec Widevine DRM. Un fichier CDM est nécessaire pour déchiffrer le contenu. Recherche "ready to use cdms" sur Google pour en trouver un.

Place le fichier `.wvd` dans le même dossier que le binaire, **ou** les fichiers `client_id.bin` et `private_key.pem`.
