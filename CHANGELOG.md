# Changelog

All notable changes to Noted will be documented in this file.

## [1.2.0] - 2024-11-29

### Added
- **Unified Trash Page** - Chrome-style trash management with category tabs (Notes, Todos, Accounts, Contacts)
- **Soft Delete for Contacts** - Contacts now go to trash instead of being permanently deleted
- **Manual Internal/External Toggle** - Mark any contact as internal or external manually
- **Bulk Delete Contacts** - Select and delete multiple contacts at once
- **4 New Themes**:
  - Dracula - Popular purple dark theme
  - Solarized - Classic developer color scheme
  - Ocean - Deep cyan aquatic theme
  - Forest - Deep emerald nature theme

### Removed
- Enterprise theme
- Liquid Glass theme

### Fixed
- TypeScript errors in contacts page
- Missing Wails routes for account trash operations
- Various UI overflow issues

## [1.1.0] - 2024-11-28

### Added
- Native macOS app using Wails framework
- Apple Calendar integration via EventKit
- Contacts feature with auto-extraction from meetings
- Domain-based account suggestions for contacts
- Multiple theme support (Modern, Minimal, Nordic, Noir, Cyber, Monokai, Retro)
- Calendar views: Month, Week, Agenda
- Data export/import functionality
- Custom delete confirmation modals (Wails compatible)

### Fixed
- Cross-platform CI/CD with build constraints
- Various TypeScript and accessibility warnings

## [1.0.0] - 2024-11-27

### Added
- Initial release
- Notes management with rich text editor (TipTap)
- Accounts/customer tracking
- Todos with Kanban board
- Full-text search (SQLite FTS4)
- Tags and templates
- PDF export
- Dark/light mode
