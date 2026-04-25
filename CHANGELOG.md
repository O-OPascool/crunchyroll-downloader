# Changelog

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
