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

---

## [Pascool-4] — 2026-04-25

### Corrections
- **Taille de police corrigée** : le PlayRes était en 640×360 dans les fichiers CR mais les styles étaient calibrés pour 1920×1080, ce qui rendait les sous-titres énormes. Le PlayRes est maintenant forcé à 640×360 avec les valeurs exactes du modèle (23px, marges 20).
- **Tag renommé** : `sipha` → `Pascool` (défaut et partout dans les noms de fichiers)

---

## [Pascool-5] — 2026-04-25

### Corrections
- **PSSH not found corrigé** : certains épisodes CR placent le PSSH dans un `AdaptationSet` différent du premier, ou directement dans une `Representation`. La recherche parcourt maintenant tout le manifeste (tous les `Period`, `AdaptationSet` et `Representation`) au lieu de s'arrêter au premier bloc.
- Au lieu de planter avec `panic`, l'épisode est maintenant **ignoré avec un warning** `⚠` si le PSSH reste vraiment introuvable, et le téléchargement continue avec l'épisode suivant.

---

## [Pascool-6] — 2026-05-06

### Corrections
- **Invalid URL format corrigé** : le code n'acceptait que les IDs de 9 ou 14 caractères. CR utilise maintenant des IDs de longueurs variables (ex: `GT00365559` = 10 chars). La vérification accepte maintenant tout ID entre 9 et 20 caractères.
