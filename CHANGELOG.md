# Changelog

## [Pascool-5] — 2026-04-25

### Corrections
- **Crash « Failed to get audio URL »** : les AdaptationSets vidéo/audio ne sont plus cherchés par index fixe (`[0]`/`[1]`) — un lookup dynamique par `MimeType`/`ContentType` est utilisé, avec fallback sur les propriétés des Representations
- **Crash « PSSH not found »** : la recherche PSSH explore maintenant les ContentProtection au niveau **Representation** en plus du niveau **AdaptationSet**
- **PSSH PlayReady au lieu de Widevine** : `getPssh` filtre désormais par le `schemeIdUri` Widevine (`edef8ba9-...`) pour ne pas retourner un PSSH PlayReady par erreur
- **Crash « nil SegmentTemplate »** : `downloadParts` résout le SegmentTemplate au niveau AdaptationSet puis Representation, au lieu de supposer qu'il est toujours au niveau AdaptationSet
- **Support des manifestes SegmentBase** : ajout de `downloadWhole` pour les épisodes dont le MPD ne contient pas de SegmentTemplate (fichier unique via BaseURL)
- **Crash silencieux sur manifest HTTP error** : `parseManifest` utilise maintenant `DoRequest` (avec refresh token 401) et vérifie le code HTTP au lieu de parser silencieusement une réponse d'erreur
- **Crash en chaîne sur batch download** : `downloadEpisode` retourne une erreur au lieu de `panic`/`os.Exit` — un épisode en erreur est sauté et le téléchargement continue

### Améliorations
- **Fallback qualité audio** : si la qualité demandée (ex: `192k`) n'existe pas dans le manifest, la meilleure qualité disponible est utilisée automatiquement (avec un avertissement affiché)
- **Recherche PSSH élargie** : tous les AdaptationSets et toutes les Representations sont inspectés, plus seulement le premier

---

## [Pascool-4] — 2026-04-25

### Corrections
- **Crash JSON `Episode.error`** : l'API Crunchyroll peut renvoyer le champ `error` sous forme de nombre au lieu d'une string — le type est maintenant `json.RawMessage` pour accepter n'importe quel type
- **Crash sur URL mal formatée** : `processUrl` ne panic plus si l'URL contient moins de 5 segments
- **Crash sur épisode introuvable** : `getEpisodeInfo` retourne une erreur propre au lieu de panic sur un tableau `data` vide
- **Crash sur saison vide** : `downloadSeason` vérifie que la liste d'épisodes n'est pas vide avant d'y accéder
- **Boucle infinie sur token invalide** : `DoRequest` limitée à 1 retry en cas de 401, au lieu d'une récursion infinie
- **Erreur silencieuse dans `token.go`** : l'erreur de `io.ReadAll` était ignorée (variable `err` shadowed) — maintenant correctement vérifiée

### Builds
- Binaires cross-compilés : `crdl-windows.exe`, `crdl-macos-arm64`, `crdl-macos-intel`, `crdl-linux`

---

## [Pascool-3] — 2026-04-25

### Améliorations
- **Restyling sous-titres entièrement revu** à partir d'un modèle fansub de référence :
  - Police **Trebuchet MS 66px** (calibrée pour 1920×1080, identique aux fansubs français professionnels)
  - Contour noir 3px + ombre 3px pour une lisibilité maximale
  - Marges 75px (plus d'espace par rapport au bord)
  - Styles complets : `Default`, `Italique`, `TiretsDefault`, `TiretsItalique`, `Sign`
  - Les panneaux/textes à l'écran (`Sign`, `Titre`, `Caption`) conservent leur style dédié (Arial, fond sombre)
  - Les noms de styles CR originaux sont préservés (les événements continuent de fonctionner)
  - `ScaledBorderAndShadow: yes` forcé pour un rendu correct sur tous les players

---

## [Pascool-2] — 2026-04-25

### Ajouts
- **Vitesse de téléchargement en MB/s** affichée en temps réel dans la barre de progression
- **Restyling automatique des sous-titres ASS** : police Arial 54px gras, contour noir épais, ombre légère, marges propres — rendu proche d'un fansub professionnel
- **README complet** avec toutes les commandes, exemples, et tableau des langues

### Améliorations
- Workers de téléchargement portés de **10 à 16** (+60% de débit sans risque de ban)
- La barre de progression track maintenant les **octets téléchargés** (pas juste le nombre de segments)

---

## [Pascool-1] — 2026-04-25

### Ajouts
- **Nouveau format de nommage** : `Title.S01E01.CR.WEBDL.VOSTFR.1080p.x265-Pascool.mkv`
  - Détection automatique `Multi` / `VOSTFR` / `VF` selon les pistes disponibles
- **Reprise automatique (resume)** : les segments sont mis en cache dans `.crdl_cache/` et survivent à une interruption
- **Barre de progression propre** avec vitesse en seg/s
- **Flag `--tag`** pour personnaliser le nom de release (défaut : `Pascool`)
- Sous-titres embarqués avec langue et titre de piste correctement renseignés (`-disposition:s:0 default`)
- Métadonnées MKV enrichies : titre, série, saison, numéro d'épisode

### Corrections
- Suppression de la fonction `sanitize` dupliquée (maintenant `sanitizeForFS` dans `output.go`)
- Affichage `⏭` propre pour les épisodes déjà téléchargés


---

## [original] — CuteTenshii

- Téléchargement d'épisodes et saisons Crunchyroll
- Déchiffrement Widevine DRM
- Support `.wvd` et `client_id.bin` + `private_key.pem`
- 10 workers parallèles
- Retry avec backoff sur erreur réseau
- Batch download via fichier texte
