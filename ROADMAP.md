# Roadmap

> Dernière mise à jour : 2026-01-08 (Tests unitaires et documentation API)

## Vision

File Viewer est un serveur HTTP local qui rend les fichiers Markdown, JSON, HTML et texte avec une interface web moderne, conçu pour l'intégration avec le panneau navigateur d'iTerm2.

## Étapes de développement

### Phase actuelle - Fonctionnalités de base

- [x] **[Feature]** Rendu Markdown avec TOC et syntax highlighting
- [x] **[Feature]** Support KaTeX pour les formules mathématiques
- [x] **[Feature]** Support Mermaid pour les diagrammes
- [x] **[Feature]** Rendu JSON interactif avec expand/collapse
- [x] **[Feature]** Mode sombre
- [x] **[Feature]** Live reload des fichiers
- [x] **[Feature]** Lightbox pour les images
- [x] **[Feature]** Recherche dans le contenu texte et JSON

### Prochaines étapes

- [x] **[Feature]** Explorateur de fichiers en sidebar (navigation dans le répertoire courant)
- [x] **[Feature]** Favoris avec persistance localStorage (groupés par répertoire)
- [x] **[Feature]** Split panels (jusqu'à 4 panneaux indépendants)
- [x] **[Feature]** Resize des panneaux par drag
- [x] **[Feature]** Support de nouveaux formats (YAML, TOML, CSV)
- [x] **[Feature]** Support des footnotes Markdown
- [x] **[Feature]** Historique de navigation des fichiers récents
- [x] **[Perf]** Mise en cache locale des dépendances CDN

### Idées et améliorations futures

- [ ] **[Feature]** Mode d'édition inline pour les fichiers
- [x] **[Feature]** Export PDF des documents Markdown
- [x] **[Feature]** Prévisualisation des liens internes au survol
- [x] **[Feature]** Support des diagrammes PlantUML
- [x] **[Feature]** Thèmes personnalisables
- [x] **[Chore]** Tests unitaires pour les fonctions de rendu
- [x] **[Docs]** Documentation API complète

## Historique des versions

### v1.10.0 - 2026-01-08
- Tests unitaires complets pour toutes les fonctions de rendu
- Tests pour Markdown, JSON, YAML, TOML, CSV, PlantUML
- Tests pour les helpers (slugify, replaceEmojis, parseCSVLine)
- Benchmarks de performance
- Documentation API complète (API.md)

### v1.9.0 - 2026-01-08
- Support des diagrammes PlantUML via ```plantuml ou ```puml
- Rendu via le serveur PlantUML officiel (plantuml.com)
- Encodage PlantUML intégré (zlib + base64 custom)

### v1.8.0 - 2026-01-08
- 6 thèmes personnalisables : Light, Dark, Sepia, Nord, Solarized Light, Solarized Dark
- Sélecteur de thème dans le header
- Persistance du thème choisi via localStorage
- Migration automatique depuis l'ancien mode sombre

### v1.7.0 - 2026-01-08
- Prévisualisation des liens internes au survol
- Popup avec aperçu du contenu du fichier lié
- Cache des prévisualisations pour performance
- Endpoint `/preview/` pour récupérer le contenu

### v1.6.0 - 2026-01-08
- Bouton d'export PDF / impression dans le header
- Styles CSS @media print optimisés pour l'impression A4
- Masquage automatique de la sidebar, TOC et toolbars à l'impression

### v1.5.0 - 2026-01-08
- Mise en cache locale des dépendances CDN (Prism.js, KaTeX, Mermaid)
- Endpoint `/cdn/` qui proxy et cache les ressources
- Cache stocké dans `~/.cache/file-viewer/cdn/`

### v1.4.0 - 2026-01-08
- Historique des fichiers récents avec section collapsible
- Suppression individuelle ou globale de l'historique
- Limite à 15 fichiers récents

### v1.3.0 - 2026-01-08
- Support des footnotes Markdown avec références et back-links

### v1.2.0 - 2026-01-08
- Support YAML avec syntax highlighting
- Support TOML avec syntax highlighting
- Support CSV avec tableau interactif et filtre

### v1.1.0 - 2026-01-08
- Sidebar avec explorateur de fichiers
- Favoris persistants groupés par répertoire
- Split panels (jusqu'à 4) avec navigation indépendante
- Resize des panneaux par drag & drop
- Filtrage des fichiers binaires et volumineux (>5MB)

### v1.0.0 - 2026-01-08
- Rendu Markdown complet (TOC, syntax highlighting, math, Mermaid)
- Rendu JSON interactif
- Mode sombre et live reload
- Intégration iTerm2
