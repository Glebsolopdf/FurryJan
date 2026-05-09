# Furryjan — E621 Content Downloader
⚠️The program is still rough and requires further corrections! It may load the processor and RAM!

![Preview](forreadme/preview.png)

A powerful CLI utility written in Go for downloading content from [e621.net](https://e621.net) by tags.  
Fully interactive interface — no flags needed, everything through menus and helpful prompts.  
Features pagination, deduplication via SQLite, tag-based folder organization, and ZIP archiving.

---

## Requirements

- **Binary**: Pre-built binary (no dependencies required after installation)
- **From Source**: Go 1.21+, GCC (for `go-sqlite3` compilation)
- **Account**: e621.net account + API key

---

## Installation

### Quick Install (Linux)

```bash
chmod +x install.sh && ./install.sh
```

The script will:
- Build the binary
- Install it to `/usr/bin/furryjan`
- Copy translations to `/usr/share/furryjan/locales`
- Create a desktop entry for your application menu

### Manual Build

```bash
cd src/
go build -o furryjan ./cmd/main.go
sudo mv furryjan /usr/bin/
```

---

## Usage

```bash
furryjan
```

Simply run it — you'll enter the main menu. No arguments or flags needed.

---

## First Run

On first launch, a setup wizard automatically appears:

```
╔══════════════════════════════════════════════╗
║                                              ║
║            Welcome to Furryjan!              ║
║                                              ║
║   Before we start, please provide the        ║
║   necessary data for the downloader to       ║
║   work properly                              ║
║                                              ║
║            (Ctrl+C to exit)                  ║
║                                              ║
╚══════════════════════════════════════════════╝

Step 0/4  Choose your language:
 1) English
 2) Русский
 > 1

Step 1/4  Enter your username on e621:
 > your_username

Step 2/4  Enter your API key (Open Profile Settings on e621 → Manage API Access):
 > ****************************

Step 3/4  Download folder (Enter = /home/user/Downloads/Furryjan):
 >

✓ Config saved: /home/user/.config/furryjan/config.json
```

Your API key is stored locally. You can find it in your e621 account settings.

---

## Main Menu

```
═════════════════════════════════════════════
        Furryjan  v1.0
═════════════════════════════════════════════
   1.  Download by Tags
   2.  Download History
   3.  Archive
   4.  Settings
   5.  Exit
═════════════════════════════════════════════
Choose:
```

---

## Features

### 1. Download by Tags

```
Search Type Selection:
 1) By Tags (regular search)
 2) Popular (order:hot)
 3) Latest (order:latest)
 4) High-Rated (order:score)
 > 1

Enter tags separated by space: dragon
Post limit (Enter = unlimited): 100
Dry-run mode? (show without downloading) [y/N]: n

→ Searching for posts with tags: dragon
→ Found: 243 posts. Starting download...

[████████████░░░░░░░░]  112/243  dragon/6378597.png  2.4 MB/s
```

**File Organization:**
- Files are saved to: `<download_dir>/<first_tag>/[post_id].[ext]`
- Already downloaded posts are skipped automatically by `post_id`
- No duplicates thanks to SQLite deduplication

### 2. Download History

Browse your last 50 downloads with timestamps and file sizes.

**Filter by Tag:**
- Search through your download history
- See all posts downloaded with a specific tag

**Statistics:**
- Total files downloaded
- Total storage used
- Number of unique tags
- First and last download dates

### 3. Archive Creator

Create ZIP archives from your downloads:
- Archive all downloads
- Archive specific tags
- Automatic compression with progress tracking

### 4. Settings

Customize your experience:

| Setting | Options | Default |
|---------|---------|---------|
| **Language** | English, Русский | English |
| **File Types** | Images only, Images + Animations, All types, Videos only | All types |
| **Log Level** | DEBUG, INFO, WARN, ERROR | INFO |
| **Blob Writer** | On/Off (optimized disk writing) | On |
| **Auto Cleanup** | Delete blob files after download | On |
| **Max File Size** | 0 = unlimited (in MB) | 0 |
| **Buffer Size** | 100-2000 MB | 500 |
| **Download Directory** | Custom path | ~/Downloads/Furryjan |

---

## Technical Details

### Storage

- **Config**: `~/.config/furryjan/config.json`
- **Database**: `~/.local/share/furryjan/furryjan.db` (SQLite)
- **Downloads**: Configurable (default: `~/Downloads/Furryjan`)

### Performance

- **Blob Writer**: Aggregates small files into larger blobs to reduce SSD write amplification
- **Buffer**: Configurable buffer size (100-2000 MB) for optimal performance
- **Deduplication**: SQLite-based database prevents duplicate downloads
- **Concurrent Downloads**: Smart parallel downloading with progress tracking

### Languages Supported

- English (en)
- Russian (Русский) (ru)

Locales are **embedded directly in the binary**, so no external files needed.

---

## Uninstall

### Using Setup Wizard

At the language selection screen, enter:
```
.delete
```

The program will guide you through safe uninstallation.

### Manual Uninstall

```bash
sudo rm /usr/bin/furryjan
sudo rm -rf ~/.config/furryjan
sudo rm -rf ~/.local/share/furryjan
```

---

## Troubleshooting

### "Failed to load translations"
- Locales are embedded in the binary, this is usually not an issue
- Try reinstalling: `./install.sh`

### Database locked error
- Close all other Furryjan instances
- Check disk space

### Download fails
- Verify your API key in settings
- Check your e621 account status
- Ensure you have write permissions to the download directory

---

## Architecture

### Project Structure

```
src/
├── cmd/
│   └── main.go                  # Entry point: init config/db, run UI
├── i18n/
│   └── locales/
│       ├── en.json              # English translations (embedded)
│       └── ru.json              # Russian translations (embedded)
└── internal/
    ├── ui/                      # Terminal UI & menu system
    │   ├── menu.go              # Main menu navigation
    │   ├── download.go          # Download flow & dialogs
    │   ├── history.go           # History & statistics views
    │   ├── archive.go           # Archive creation interface
    │   ├── settings.go          # Settings management UI
    │   └── render.go            # UI helpers (boxes, colors, input)
    ├── api/                     # E621 API client
    │   ├── client.go            # HTTP client with rate limiting
    │   ├── types.go             # API response structures
    │   └── manager.go           # API request management
    ├── config/                  # Configuration management
    │   ├── config.go            # Config loading/saving
    │   ├── setup.go             # Initial setup wizard
    │   ├── selfinstall.go       # Binary self-installation
    │   └── uninstall.go         # Uninstall handler
    ├── db/                      # SQLite database
    │   ├── db.go                # Database initialization
    │   ├── downloads.go         # Download history tracking
    │   └── stats.go             # Statistics calculation
    ├── downloader/              # Download engine
    │   ├── downloader.go        # Main download loop
    │   ├── file.go              # Individual file downloads
    │   ├── progress.go          # Progress tracking
    │   └── blob/                # Blob writer optimization
    │       ├── writer.go        # Aggregated blob writing
    │       ├── manager.go       # Blob lifecycle management
    │       ├── extract.go       # Blob extraction & cleanup
    │       └── doc.go           # Documentation
    └── archiver/                # ZIP archive creation
        ├── archiver.go          # Archive creation logic
        └── filter.go            # Tag filtering for archives
```

### Data Flow

```
User Input (Menu)
    ↓
config (Load settings)
    ↓
api.GetPosts (E621 API)
    ↓
downloader.Run (Main loop)
    ├─ Fetch posts (paginated)
    ├─ Check db (IsDownloaded)
    ├─ Download files (with progress)
    ├─ Save to blob/filesystem
    └─ Store in db (SaveDownload)
    ↓
database (SQLite)
    ├─ Track downloads
    ├─ Store statistics
    └─ Enable deduplication
    ↓
User sees results
```

### Key Components

| Component | Purpose |
|-----------|---------|
| **UI Package** | Terminal interface with menu-driven navigation |
| **API Package** | E621 API client with rate limiting & auth |
| **Config Package** | User settings and application configuration |
| **DB Package** | SQLite database for download tracking |
| **Downloader** | Core download engine with blob optimization |
| **Archiver** | ZIP file creation with filtering |

### Design Principles

1. **No circular dependencies** - All packages are loosely coupled
2. **CLI-first** - Interactive menus, no command-line flags
3. **Offline-capable** - History and statistics work without internet
4. **Efficient storage** - Blob writer reduces SSD write amplification
5. **Auto-deduplication** - SQLite prevents duplicate downloads
6. **Multi-language** - Embedded translations (English & Russian)

---

## E621 API Documentation

### Authentication

Furryjan uses **Basic Auth** to authenticate with the E621 API:

```
Authorization: Basic <base64(username:api_key)>
```

**Getting your API key:**
1. Go to [e621.net](https://e621.net)
2. Login to your account
3. Navigate to: Account → My Profile → Manage API Access
4. Generate or copy your API key
5. Enter it in Furryjan's setup wizard

### Rate Limiting

E621 enforces a **rate limit of 2 requests per second**:

```
Request Rate: ≤ 2 requests/second (hard limit)
Recommended: ≤ 1 request/second (best practice)
Timeout: 503 Service Unavailable if exceeded
```

Furryjan automatically handles rate limiting to avoid exceeding these limits.

### API Endpoints Used

| Endpoint | Purpose | Example |
|----------|---------|---------|
| `/posts.json` | Fetch posts by tags | `?tags=dragon&limit=320&page=1` |
| `/posts/{id}.json` | Get single post | Fetch post details |
| `/posts/search/similar.json` | Similar posts | Not currently used |

### Query Parameters

When searching for posts, you can use:

**Basic Parameters:**
- `tags` - Space-separated search terms (e.g., `dragon fox male`)
- `limit` - Number of posts per page (max: 320)
- `page` - Page number for pagination

**Search Modifiers:**
- `order:hot` - Most popular posts
- `order:latest` - Newest posts first
- `order:score` - Highest rated posts
- `order:random` - Random posts

**File Filters:**
- `type:png`, `type:jpg`, `type:gif` - File types
- `-animated` - Exclude animated/video
- `animated` - Only animated content

### Response Format

E621 API returns JSON with post data:

```json
{
  "id": 6378597,
  "created_at": "2024-01-15T10:30:00.000Z",
  "updated_at": "2024-01-15T10:35:00.000Z",
  "file": {
    "width": 1920,
    "height": 1080,
    "ext": "png",
    "size": 2097152,
    "md5": "abcdef123456",
    "url": "https://static1.e621.net/data/abcdef123456.png"
  },
  "preview": {
    "width": 150,
    "height": 150,
    "url": "https://static1.e621.net/data/preview/abcdef123456.jpg"
  },
  "sample": {
    "width": 320,
    "height": 180,
    "url": "https://static1.e621.net/data/sample/abcdef123456.jpg"
  },
  "score": 42,
  "tags": {
    "general": ["dragon", "male"],
    "species": ["dragon"],
    "character": [],
    "copyright": [],
    "artist": ["artist_name"],
    "invalid": [],
    "lore": [],
    "meta": []
  },
  "locked": {
    "notes": false,
    "rating": false,
    "status": false
  },
  "rating": "q",
  "fav_count": 123,
  "comment_count": 5,
  "has_children": false
}
```

### Important Notes

⚠️ **User-Agent Required:**
- All requests must include a descriptive User-Agent header
- Furryjan uses: `Furryjan (by username on e621)`
- Never impersonate browser user agents (will be blocked)

⚠️ **Rate Limit Policy:**
- Hard limit: 2 requests/second
- Recommended: ≤ 1 request/second
- Sustained exceeding will result in IP bans

⚠️ **Copyright & Terms:**
- Respect content creators' rights
- Follow e621 Terms of Service
- Don't abuse the API for commercial purposes

### For More Information

- **E621 API Docs**: https://e621.net/help/api
- **E621 Tags**: https://e621.net/tags
- **E621 Forum**: https://e621.net/forum

---

## Credits

Built with:
- **Go 1.21+** - Fast, efficient binary
- **SQLite** - Reliable local database
- **go-sqlite3** - Database driver
- **progressbar/v3** - Download progress visualization
- **E621 API** - Content source

---

## License

MIT License - See [LICENSE](LICENSE) file for details

---

## Support

For issues, suggestions, or improvements:
1. Check the [existing documentation](src/api.md)
2. Verify your configuration in settings
3. Review your API key permissions

![Enjoy](forreadme/enjoy.png)
