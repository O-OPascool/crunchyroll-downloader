# crunchyroll-downloader

**[🇫🇷 Français](#-français)** | **[🇬🇧 English](#-english)**

---

## 🇫🇷 Français

Télécharge des animes depuis Crunchyroll et les exporte en fichiers `.mkv`.

Développé par [CuteTenshii](https://github.com/CuteTenshii/crunchyroll-downloader) — amélioré par **Pascool**.

---

### 💻 Plateformes supportées

* Windows
* Linux
* macOS (Intel & Apple Silicon ARM)

---

### Prérequis

* [FFmpeg](https://www.ffmpeg.org/download.html) — doit être accessible dans le PATH ou dans le même dossier
* Un fichier `.wvd` **ou** les fichiers `client_id.bin` + `private_key.pem` (Widevine CDM)
* Un compte Crunchyroll (Premium requis pour le contenu Premium)

---

### Installation

1. Télécharge le binaire correspondant à ton OS depuis :
   https://github.com/CuteTenshii/crunchyroll-downloader/releases/latest

2. Place `ffmpeg` dans le même dossier ou dans le PATH

3. Ajoute ton fichier `.wvd` (ou `client_id.bin` + `private_key.pem`)

4. Récupère ton cookie `etp_rt`

---

### Authentification

Le cookie `etp_rt` est nécessaire.

**Étapes :**

1. Va sur https://www.crunchyroll.com
2. Connecte-toi
3. Ouvre les DevTools (`Ctrl+Shift+I`)
4. Onglet **Application / Storage → Cookies**
5. Copie la valeur du cookie `etp_rt`

**Exemple :**

![etp\_rt cookie screenshot](https://raw.githubusercontent.com/O-OPascool/crunchyroll-downloader/master/.github/screenshots/etp-rt-cookie.png)

---

### Utilisation

```
crdl-windows.exe [options]
```

**Options :**

```
-url string
-urls string
-etp-rt string
-season int
-audio-lang string
-subs-lang string
-video-quality string
-audio-quality string
-tag string
```

---

### Exemples

**Épisode unique**

```bash
crdl-windows.exe --url https://www.crunchyroll.com/watch/... --etp-rt TON_COOKIE
```

**Saison complète**

```bash
crdl-windows.exe --url https://www.crunchyroll.com/series/... --season 1 --etp-rt TON_COOKIE
```

---

### Format de sortie

```
SeriesTitle.S01E01.CR.WEBDL.VOSTFR.1080p.x265-Pascool.mkv
```

---

## 🇬🇧 English

Downloads anime from Crunchyroll and outputs `.mkv` files.

---

### 💻 Supported platforms

* Windows
* Linux
* macOS (Intel & Apple Silicon ARM)

---

### Authentication

The `etp_rt` cookie is required.

**Steps:**

1. Go to https://www.crunchyroll.com
2. Log in
3. Open DevTools (`Ctrl+Shift+I`)
4. Go to Cookies
5. Copy `etp_rt`

**Example:**

![etp\_rt cookie screenshot](https://raw.githubusercontent.com/O-OPascool/crunchyroll-downloader/master/.github/screenshots/etp-rt-cookie.png)

---

### Usage

```
crdl-windows.exe [options]
```
