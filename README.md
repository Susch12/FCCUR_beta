# FCCUR - Free Community Content Universal Repository

> **Sistema Local de Distribución de Herramientas Educativas**
> Un servidor pragmático de paquetes educativos diseñado para entornos universitarios con conectividad limitada.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Raspberry%20Pi%204-C51A4A?style=flat&logo=raspberry-pi)](https://www.raspberrypi.org/)
[![Database](https://img.shields.io/badge/Database-SQLite%20%7C%20PostgreSQL-blue)](https://www.sqlite.org/)

---

## 🌱 Plant This Seed Anywhere

**FCCUR is now fully replicable** - a true "seed" that can be planted on any system:

- ✅ **Zero Configuration**: Works out of the box with sensible defaults
- ✅ **Database Agnostic**: Auto-detects SQLite or PostgreSQL
- ✅ **Environment Variables**: Full configuration via env vars
- ✅ **Production Ready**: Secure JWT, TLS/HTTPS, OAuth2 support
- ✅ **Migration System**: Automated database versioning with rollback
- ✅ **Cross-Platform**: Linux, macOS, Windows (WSL), ARM (Raspberry Pi)

```bash
# Clone, build, run - that's it!
git clone <repo-url>
cd FCCUR
make build-all
./bin/fccur
```

**Deploy anywhere in < 5 minutes** → See [DEPLOYMENT.md](DEPLOYMENT.md) for details.

---

## 📖 Tabla de Contenidos

- [🌱 Plant This Seed Anywhere](#-plant-this-seed-anywhere)
- [🚀 Quick Start](#-quick-start)
- [✨ What's New](#-whats-new)
- [Problema que Resuelve](#-problema-que-resuelve)
- [Solución](#-solución)
- [Características Principales](#-características-principales)
- [Arquitectura Técnica](#-arquitectura-técnica)
- [Instalación y Setup](#-instalación-y-setup)
- [Configuración](#-configuración)
- [Database Options](#-database-options)
- [Migrations](#-migrations)
- [Uso](#-uso)
- [API Reference](#-api-reference)
- [Deployment](#-deployment)
- [Roadmap](#-roadmap)
- [Contribuir](#-contribuir)

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ ([Download](https://go.dev/dl/))
- GCC (for SQLite CGO)
- Git

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/FCCUR.git
cd FCCUR

# Build binaries
make build-all

# Run server (SQLite, port 8080)
./bin/fccur
```

Server starts at `http://localhost:8080`

### With PostgreSQL

```bash
# Install PostgreSQL
sudo apt install postgresql

# Create database
sudo -u postgres createdb fccur

# Run with PostgreSQL
export FCCUR_DB="postgres://postgres:password@localhost/fccur"
./bin/fccur
```

### Configuration Options

```bash
# Environment variables (recommended)
export FCCUR_ADDR=":8080"
export FCCUR_DB="./data/fccur.db"
export FCCUR_JWT_SECRET="$(openssl rand -base64 32)"

# Or command-line flags
./bin/fccur -addr=":8080" -db="./data/fccur.db"

# Or systemd service
make deploy
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for comprehensive deployment guide.

---

## ✨ What's New

### Version 2.0 - The Replicability Update

**Major improvements for easy deployment anywhere:**

#### 🔧 Configuration Flexibility
- **Environment Variable Support**: All settings configurable via `FCCUR_*` env vars
- **Configurable Paths**: No more hardcoded paths - web, packages, migrations all configurable
- **Smart Defaults**: Works immediately with zero configuration

#### 🗄️ Database Enhancements
- **PostgreSQL Support**: Full production-ready PostgreSQL integration with connection pooling
- **Auto-Detection**: Automatically detects SQLite vs PostgreSQL from connection string
- **Migration System**: golang-migrate integration with version tracking
  - Up/down migrations
  - Safe rollback capability
  - CLI tool for migration management
- **Dual Database**: Keep SQLite for development, PostgreSQL for production

#### 🔐 Security Improvements
- **Proper JWT Signing**: Replaced weak implementation with HMAC-SHA256
- **No Hardcoded Secrets**: All secrets via environment variables
- **Production Ready**: TLS/HTTPS, OAuth2, rate limiting all configured

#### 📦 New Features
- **Migration CLI**: `bin/migrate` for database version management
- **Health Checks**: `/health` endpoint for monitoring
- **Better Logging**: Request logging with timing information
- **Makefile Targets**: `make migrate-up`, `make migrate-down`, etc.

#### 📚 Documentation
- [DEPLOYMENT.md](DEPLOYMENT.md) - Comprehensive deployment guide
- [MIGRATIONS.md](MIGRATIONS.md) - Database migration documentation
- [DATABASE_ABSTRACTION.md](DATABASE_ABSTRACTION.md) - Architecture details

### Environment Variables

All configuration now supports environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `FCCUR_ADDR` | `:8080` | Server address |
| `FCCUR_DB` | `./data/fccur.db` | Database connection string |
| `FCCUR_PACKAGES_DIR` | `./packages` | Packages directory |
| `FCCUR_WEB_DIR` | `./web` | Web files directory |
| `FCCUR_MIGRATIONS_DIR` | `./migrations` | Migrations directory |
| `FCCUR_JWT_SECRET` | (auto-generated) | JWT signing secret |
| `FCCUR_CERT_FILE` | - | TLS certificate (optional) |
| `FCCUR_KEY_FILE` | - | TLS private key (optional) |
| `FCCUR_AUTH_USER` | - | Upload auth username (optional) |
| `FCCUR_AUTH_PASS` | - | Upload auth password (optional) |
| `FCCUR_RATE_LIMIT` | `10` | Uploads per hour per IP |
| `FCCUR_OAUTH2_CLIENT_ID` | - | OAuth2 client ID (optional) |
| `FCCUR_OAUTH2_CLIENT_SECRET` | - | OAuth2 client secret (optional) |
| `FCCUR_OAUTH2_REDIRECT_URL` | `http://localhost:8080/api/oauth2/callback` | OAuth2 redirect URL |
| `FCCUR_OAUTH2_TENANT` | `common` | Microsoft tenant ID |

---

## 🎯 Problema que Resuelve

### Situación Actual en Universidades

**Pain Points**:
- ⏱️ **2-4 horas** de setup por estudiante para instalar herramientas de desarrollo
- 🐌 **Internet lento** colapsa con 30+ estudiantes descargando simultáneamente
- 🔄 **Configuraciones inconsistentes** entre máquinas causan problemas de compatibilidad
- 📚 **Primera semana** del curso dedicada a instalaciones en lugar de aprendizaje

### Ejemplo Real
```
Curso: Sistemas Operativos II
Herramientas: Ubuntu 22.04 (4.7GB), VirtualBox (150MB), GCC (200MB), VS Code (85MB)

Tiempo actual por estudiante:
- Descarga: 90-120 min (internet lento)
- Instalación: 45-60 min
- Resolución de problemas: 30-90 min
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total: 2.5 - 4.5 horas

Con 30 estudiantes:
- Tiempo perdido: 75-135 horas
- Tráfico de red: 156 GB (¡saturando la red!)
```

---

## ✅ Solución

**FCCUR** es un servidor local que:

1. 🚀 **Reduce el tiempo de setup** de 2+ horas a **<15 minutos**
2. 🏠 **Funciona completamente offline** en red local (Gigabit LAN)
3. 🔒 **Garantiza versiones estandarizadas** para todos los estudiantes
4. 💰 **Costo ultra-bajo**: $115 USD en hardware (Raspberry Pi 4)

### Antes vs Después

```
┌─────────────────────┐         ┌──────────────────────┐
│   ANTES (Internet)  │         │  DESPUÉS (FCCUR)     │
├─────────────────────┤         ├──────────────────────┤
│ ⏱️  2+ horas        │   →     │ ⏱️  <15 minutos      │
│ 🐌 Dependiente ISP  │   →     │ 🚀 Gigabit LAN       │
│ ❌ 156GB tráfico    │   →     │ ✅ 1 descarga/30     │
│ 🔄 Versiones mixtas │   →     │ 🔒 100% consistente  │
└─────────────────────┘         └──────────────────────┘
```

---

## 🌟 Características Principales

### Core Features (Production Ready)
- ✅ **Servidor HTTP en Go** (stdlib, zero frameworks)
- ✅ **Dual Database Support** (SQLite + PostgreSQL with auto-detection)
- ✅ **Database Migrations** (golang-migrate with version tracking)
- ✅ **Dual Hashing** (BLAKE3 + SHA256)
- ✅ **API REST completa** (upload, download, list, stats)
- ✅ **Web UI responsive** (vanilla JS, offline-capable)
- ✅ **File streaming optimizado**
- ✅ **Download statistics tracking**
- ✅ **Environment Variable Configuration** (12-factor app compliant)
- ✅ **Secure JWT Authentication** (HMAC-SHA256)
- ✅ **OAuth2 Integration** (Microsoft/Azure AD)
- ✅ **TLS/HTTPS Support**
- ✅ **Rate Limiting** (configurable per-IP limits)
- ✅ **Health Checks** (monitoring endpoint)
- ✅ **Request Logging** (with timing information)

### Database & Storage
- ✅ **SQLite** for development and edge deployments
- ✅ **PostgreSQL** for production with connection pooling
- ✅ **Automatic schema migrations** on startup
- ✅ **Safe rollback capability** via CLI tool
- ✅ **BLAKE3 deduplication** preventing duplicate uploads
- ✅ **WAL mode** for better SQLite concurrency

### Security & Authentication
- ✅ **HMAC-SHA256 JWT signing** (production-grade)
- ✅ **OAuth2 support** (Microsoft/Azure AD ready)
- ✅ **Basic auth** for upload protection
- ✅ **Rate limiting** to prevent abuse
- ✅ **Privacy-first headers** (no-referrer, DNS prefetch control)
- ✅ **CORS support** for cross-origin requests
- ✅ **Session management** with refresh tokens

### DevOps & Deployment
- ✅ **Single binary** (no external dependencies)
- ✅ **Cross-platform** (Linux, macOS, Windows, ARM)
- ✅ **Systemd service** integration
- ✅ **Docker ready** (configurable via env vars)
- ✅ **Makefile automation** (build, deploy, migrate)
- ✅ **Health monitoring** endpoints
- ✅ **Graceful shutdown** handling

### Future Enhancements (Roadmap)
- 🔄 **Multi-node synchronization** (CRDT)
- 🔐 **Native sandboxing** (Linux namespaces)
- 🌐 **Distributed architecture** (Raspberry Pi cluster)
- 🔑 **Web of Trust authentication**
- 📊 **Advanced analytics dashboard**
- 🐳 **Official Docker images**
- ☸️ **Kubernetes manifests**

---

## 🏗️ Arquitectura Técnica

### Stack Tecnológico

```yaml
Backend:
  Lenguaje: Go 1.21+
  Database: SQLite 3.x (development) | PostgreSQL 14+ (production)
  HTTP: net/http (stdlib)
  Hashing: BLAKE3 (primario) + SHA256 (fallback)
  Migrations: golang-migrate/migrate v4
  Connection Pooling: pgx/pgxpool (PostgreSQL)
  Authentication: JWT (HMAC-SHA256) + OAuth2

Frontend:
  HTML5: Estructura semántica
  CSS3: Responsive design (mobile-first)
  JavaScript: Vanilla ES6+ (Fetch API)

Infrastructure:
  Hardware: Raspberry Pi 4 (4GB RAM) | Any x86_64/ARM64 system
  Storage: SSD USB 3.2 (256GB) | Cloud storage
  Network: Gigabit Ethernet | Any network
  Deployment: Systemd | Docker | Kubernetes-ready
  OS: Raspberry Pi OS (Debian-based)
```

### Arquitectura del Sistema

```
┌─────────────────────────────────────────────────┐
│              RASPBERRY PI 4                     │
│  ┌───────────────────────────────────────────┐  │
│  │           FRONTEND (Static)               │  │
│  │  ┌─────────┐ ┌─────────┐ ┌────────────┐  │  │
│  │  │index.html│ │style.css│ │   app.js   │  │  │
│  │  └─────────┘ └─────────┘ └────────────┘  │  │
│  └──────────────────┬────────────────────────┘  │
│                     │ HTTP Requests             │
│  ┌──────────────────▼────────────────────────┐  │
│  │         BACKEND (Go)                      │  │
│  │  ┌──────────────────────────────────────┐ │  │
│  │  │     HTTP Server (net/http)           │ │  │
│  │  │  - Routes & Handlers                 │ │  │
│  │  │  - CORS Middleware                   │ │  │
│  │  │  - Logging                           │ │  │
│  │  └──────────┬───────────────────────────┘ │  │
│  │             │                              │  │
│  │  ┌──────────▼───────────────────────────┐ │  │
│  │  │         API Layer                    │ │  │
│  │  │  GET  /api/packages                  │ │  │
│  │  │  GET  /api/packages/:id              │ │  │
│  │  │  POST /api/packages                  │ │  │
│  │  │  GET  /download/:id                  │ │  │
│  │  │  GET  /api/stats                     │ │  │
│  │  └──────────┬───────────────────────────┘ │  │
│  │             │                              │  │
│  │  ┌──────────▼───────────────────────────┐ │  │
│  │  │      Business Logic                  │ │  │
│  │  │  - Package validation                │ │  │
│  │  │  - Dual hashing (BLAKE3+SHA256)      │ │  │
│  │  │  - File operations                   │ │  │
│  │  │  - Stats calculation                 │ │  │
│  │  └──────────┬───────────────────────────┘ │  │
│  │             │                              │  │
│  │  ┌──────────▼───────────────────────────┐ │  │
│  │  │      Storage Layer (SQLite)          │ │  │
│  │  │  Tables:                             │ │  │
│  │  │    - packages (metadata)             │ │  │
│  │  │    - downloads (stats)               │ │  │
│  │  └──────────┬───────────────────────────┘ │  │
│  └─────────────┼──────────────────────────────┘  │
│                │                                  │
│  ┌─────────────▼──────────────────────────────┐  │
│  │     File System (External SSD)            │  │
│  │     /packages/                            │  │
│  │       ├── ubuntu-22.04.iso                │  │
│  │       ├── virtualbox-7.0.exe              │  │
│  │       ├── vscode-1.84.deb                 │  │
│  │       └── ...                             │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
            │
            │ Network (Gigabit Ethernet)
            ▼
┌──────────────────────────┐
│   Estudiantes (Clientes) │
│   - Windows / Mac / Linux│
│   - Navegador web        │
│   - Scripts instalación  │
└──────────────────────────┘
```

### Modelo de Datos

```sql
-- Tabla principal: packages
CREATE TABLE packages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  description TEXT,
  category TEXT NOT NULL,
  file_path TEXT NOT NULL UNIQUE,
  file_size INTEGER NOT NULL,
  blake3_hash TEXT NOT NULL,
  sha256_hash TEXT NOT NULL,
  download_url TEXT,
  platform TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de estadísticas
CREATE TABLE downloads (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  package_id INTEGER NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  downloaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (package_id) REFERENCES packages(id)
);

-- Índices para performance
CREATE INDEX idx_packages_category ON packages(category);
CREATE INDEX idx_packages_platform ON packages(platform);
CREATE INDEX idx_downloads_package ON downloads(package_id);
CREATE INDEX idx_downloads_date ON downloads(downloaded_at);
```

---

## 🚀 Guía de Desarrollo (5 Días)

### Filosofía: "Funciona primero, optimiza después"

Este plan te lleva de **0 a MVP funcional** en 5 días de trabajo enfocado.

---

### **DÍA 1: Foundation & Core Backend** ⏱️ 8 horas

**Objetivo**: Backend funcional que puede recibir y servir archivos

#### Setup Inicial (0-2h)

```bash
# 1. Crear estructura del proyecto
mkdir -p fccur/{cmd/server,internal/{api,storage,models,hash},web,packages,data}
cd fccur

# 2. Inicializar Go module
go mod init github.com/yourusername/fccur

# 3. Instalar dependencias
go get github.com/mattn/go-sqlite3
go get github.com/zeebo/blake3

# 4. Crear .gitignore
cat > .gitignore << 'EOF'
*.db
packages/*
!packages/.gitkeep
data/*
!data/.gitkeep
fccur
fccur-*
.env
EOF

# 5. Crear directorios necesarios
touch packages/.gitkeep data/.gitkeep
```

#### Estructura de Archivos

```
fccur/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── storage/
│   │   ├── db.go               # Database operations
│   │   └── schema.go           # Schema & migrations
│   ├── api/
│   │   ├── handlers.go         # HTTP handlers
│   │   ├── routes.go           # Route setup
│   │   └── middleware.go       # CORS, logging
│   ├── models/
│   │   └── package.go          # Package struct
│   └── hash/
│       └── hash.go             # BLAKE3 + SHA256
├── web/
│   ├── index.html
│   ├── style.css
│   └── app.js
├── packages/                    # File storage
├── data/                        # SQLite DB
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

#### Implementación Core (2-8h)

**1. Database Schema** (`internal/storage/schema.go`)

```go
package storage

const schema = `
CREATE TABLE IF NOT EXISTS packages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  description TEXT,
  category TEXT NOT NULL,
  file_path TEXT NOT NULL UNIQUE,
  file_size INTEGER NOT NULL,
  blake3_hash TEXT NOT NULL,
  sha256_hash TEXT NOT NULL,
  download_url TEXT,
  platform TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS downloads (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  package_id INTEGER NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  downloaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (package_id) REFERENCES packages(id)
);

CREATE INDEX IF NOT EXISTS idx_packages_category ON packages(category);
CREATE INDEX IF NOT EXISTS idx_packages_platform ON packages(platform);
CREATE INDEX IF NOT EXISTS idx_downloads_package ON downloads(package_id);
CREATE INDEX IF NOT EXISTS idx_downloads_date ON downloads(downloaded_at);
`
```

**2. Models** (`internal/models/package.go`)

```go
package models

import "time"

type Package struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Version     string    `json:"version"`
    Description string    `json:"description,omitempty"`
    Category    string    `json:"category"`
    FilePath    string    `json:"file_path"`
    FileSize    int64     `json:"file_size"`
    BLAKE3Hash  string    `json:"blake3_hash"`
    SHA256Hash  string    `json:"sha256_hash"`
    DownloadURL string    `json:"download_url,omitempty"`
    Platform    string    `json:"platform"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type DownloadStats struct {
    PackageID      int64     `json:"package_id"`
    PackageName    string    `json:"package_name"`
    TotalDownloads int       `json:"total_downloads"`
    LastDownload   time.Time `json:"last_download,omitempty"`
}
```

**3. Dual Hashing** (`internal/hash/hash.go`)

```go
package hash

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    
    "github.com/zeebo/blake3"
)

// DualHash calcula BLAKE3 y SHA256 en un solo pass
func DualHash(r io.Reader) (blake3Hash, sha256Hash string, err error) {
    b3 := blake3.New()
    s256 := sha256.New()
    
    // Escribir a ambos hasher simultáneamente
    writer := io.MultiWriter(b3, s256)
    if _, err := io.Copy(writer, r); err != nil {
        return "", "", err
    }
    
    return hex.EncodeToString(b3.Sum(nil)),
           hex.EncodeToString(s256.Sum(nil)),
           nil
}

// VerifyBLAKE3 verifica el hash BLAKE3 de un archivo
func VerifyBLAKE3(r io.Reader, expectedHash string) (bool, error) {
    b3 := blake3.New()
    if _, err := io.Copy(b3, r); err != nil {
        return false, err
    }
    
    actualHash := hex.EncodeToString(b3.Sum(nil))
    return actualHash == expectedHash, nil
}
```

**4. Main Server** (`cmd/server/main.go`)

```go
package main

import (
    "flag"
    "log"
    "net/http"
    
    "github.com/yourusername/fccur/internal/api"
    "github.com/yourusername/fccur/internal/storage"
)

func main() {
    // Flags
    addr := flag.String("addr", ":8080", "HTTP server address")
    dbPath := flag.String("db", "./data/fccur.db", "Database path")
    packagesDir := flag.String("packages", "./packages", "Packages directory")
    flag.Parse()
    
    // Inicializar database
    db, err := storage.NewDatabase(*dbPath)
    if err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }
    defer db.Close()
    
    // Ejecutar migraciones
    if err := db.Migrate(); err != nil {
        log.Fatalf("Error running migrations: %v", err)
    }
    
    // Crear servidor API
    server := api.NewServer(db, *packagesDir)
    
    // Iniciar servidor
    log.Printf("🚀 FCCUR server starting on %s", *addr)
    log.Printf("📦 Packages directory: %s", *packagesDir)
    log.Printf("💾 Database: %s", *dbPath)
    
    if err := http.ListenAndServe(*addr, server); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

#### Checklist Día 1

- [ ] Proyecto inicializado con Go modules
- [ ] Estructura de directorios creada
- [ ] SQLite schema definido
- [ ] Modelos de datos implementados
- [ ] Dual hashing funcional (BLAKE3 + SHA256)
- [ ] HTTP server básico corriendo
- [ ] Primer test manual exitoso

**Output esperado**: `curl http://localhost:8080/health` → `{"status": "ok"}`

---

### **DÍA 2: API REST Completa** ⏱️ 8 horas

**Objetivo**: Todos los endpoints funcionando

#### Database Operations (0-2h)

**`internal/storage/db.go`**

```go
package storage

import (
    "database/sql"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "github.com/yourusername/fccur/internal/models"
)

type Database struct {
    db *sql.DB
}

func NewDatabase(path string) (*Database, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }
    
    // Optimizaciones SQLite
    db.SetMaxOpenConns(1) // SQLite solo soporta 1 escritor
    db.Exec("PRAGMA journal_mode=WAL")
    db.Exec("PRAGMA synchronous=NORMAL")
    db.Exec("PRAGMA cache_size=10000")
    
    return &Database{db: db}, nil
}

func (d *Database) Close() error {
    return d.db.Close()
}

func (d *Database) Migrate() error {
    _, err := d.db.Exec(schema)
    return err
}

// CreatePackage inserta un nuevo paquete
func (d *Database) CreatePackage(pkg *models.Package) (int64, error) {
    query := `
        INSERT INTO packages (name, version, description, category, file_path,
            file_size, blake3_hash, sha256_hash, download_url, platform)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    result, err := d.db.Exec(query,
        pkg.Name, pkg.Version, pkg.Description, pkg.Category,
        pkg.FilePath, pkg.FileSize, pkg.BLAKE3Hash, pkg.SHA256Hash,
        pkg.DownloadURL, pkg.Platform,
    )
    if err != nil {
        return 0, err
    }
    
    return result.LastInsertId()
}

// GetPackage obtiene un paquete por ID
func (d *Database) GetPackage(id int64) (*models.Package, error) {
    query := `SELECT * FROM packages WHERE id = ?`
    
    pkg := &models.Package{}
    err := d.db.QueryRow(query, id).Scan(
        &pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description,
        &pkg.Category, &pkg.FilePath, &pkg.FileSize,
        &pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL,
        &pkg.Platform, &pkg.CreatedAt, &pkg.UpdatedAt,
    )
    
    return pkg, err
}

// ListPackages obtiene todos los paquetes
func (d *Database) ListPackages() ([]*models.Package, error) {
    query := `SELECT * FROM packages ORDER BY created_at DESC`
    
    rows, err := d.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    packages := []*models.Package{}
    for rows.Next() {
        pkg := &models.Package{}
        err := rows.Scan(
            &pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description,
            &pkg.Category, &pkg.FilePath, &pkg.FileSize,
            &pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL,
            &pkg.Platform, &pkg.CreatedAt, &pkg.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        packages = append(packages, pkg)
    }
    
    return packages, nil
}

// RecordDownload registra una descarga
func (d *Database) RecordDownload(packageID int64, ip, userAgent string) error {
    query := `
        INSERT INTO downloads (package_id, ip_address, user_agent)
        VALUES (?, ?, ?)
    `
    _, err := d.db.Exec(query, packageID, ip, userAgent)
    return err
}

// GetStats obtiene estadísticas de descargas
func (d *Database) GetStats() ([]*models.DownloadStats, error) {
    query := `
        SELECT 
            p.id,
            p.name,
            COUNT(d.id) as total,
            MAX(d.downloaded_at) as last_download
        FROM packages p
        LEFT JOIN downloads d ON p.id = d.package_id
        GROUP BY p.id
        ORDER BY total DESC
    `
    
    rows, err := d.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    stats := []*models.DownloadStats{}
    for rows.Next() {
        s := &models.DownloadStats{}
        var lastDownload sql.NullTime
        
        err := rows.Scan(&s.PackageID, &s.PackageName, &s.TotalDownloads, &lastDownload)
        if err != nil {
            return nil, err
        }
        
        if lastDownload.Valid {
            s.LastDownload = lastDownload.Time
        }
        
        stats = append(stats, s)
    }
    
    return stats, nil
}
```

#### HTTP Handlers (2-6h)

**`internal/api/handlers.go`**

```go
package api

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    
    "github.com/yourusername/fccur/internal/hash"
    "github.com/yourusername/fccur/internal/models"
)

// GetPackages lista todos los paquetes
func (s *Server) GetPackages(w http.ResponseWriter, r *http.Request) {
    packages, err := s.db.ListPackages()
    if err != nil {
        http.Error(w, "Error fetching packages", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(packages)
}

// GetPackage obtiene un paquete específico
func (s *Server) GetPackage(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Query().Get("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
    pkg, err := s.db.GetPackage(id)
    if err != nil {
        http.Error(w, "Package not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(pkg)
}

// UploadPackage sube un nuevo paquete
func (s *Server) UploadPackage(w http.ResponseWriter, r *http.Request) {
    // Limitar tamaño (10GB)
    if err := r.ParseMultipartForm(10 << 30); err != nil {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }
    
    // Obtener archivo
    file, header, err := r.FormFile("package")
    if err != nil {
        http.Error(w, "Error reading file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Obtener metadata
    name := r.FormValue("name")
    version := r.FormValue("version")
    category := r.FormValue("category")
    platform := r.FormValue("platform")
    description := r.FormValue("description")
    
    // Validar
    if name == "" || version == "" || category == "" {
        http.Error(w, "Missing required fields", http.StatusBadRequest)
        return
    }
    
    // Guardar archivo
    destPath := filepath.Join(s.packagesDir, header.Filename)
    dest, err := os.Create(destPath)
    if err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    defer dest.Close()
    
    // Copiar y calcular hashes
    blake3Hash, sha256Hash, fileSize, err := s.saveAndHash(file, dest)
    if err != nil {
        os.Remove(destPath)
        http.Error(w, "Error processing file", http.StatusInternalServerError)
        return
    }
    
    // Crear registro en DB
    pkg := &models.Package{
        Name:        name,
        Version:     version,
        Description: description,
        Category:    category,
        FilePath:    destPath,
        FileSize:    fileSize,
        BLAKE3Hash:  blake3Hash,
        SHA256Hash:  sha256Hash,
        Platform:    platform,
    }
    
    id, err := s.db.CreatePackage(pkg)
    if err != nil {
        os.Remove(destPath)
        http.Error(w, "Error creating package", http.StatusInternalServerError)
        return
    }
    
    pkg.ID = id
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(pkg)
}

// DownloadPackage descarga un archivo
func (s *Server) DownloadPackage(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Query().Get("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
    // Obtener metadata
    pkg, err := s.db.GetPackage(id)
    if err != nil {
        http.Error(w, "Package not found", http.StatusNotFound)
        return
    }
    
    // Abrir archivo
    file, err := os.Open(pkg.FilePath)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer file.Close()
    
    // Registrar descarga (async)
    go s.db.RecordDownload(id, r.RemoteAddr, r.UserAgent())
    
    // Headers
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(pkg.FilePath)))
    w.Header().Set("Content-Length", strconv.FormatInt(pkg.FileSize, 10))
    w.Header().Set("X-BLAKE3-Hash", pkg.BLAKE3Hash)
    w.Header().Set("X-SHA256-Hash", pkg.SHA256Hash)
    
    // Stream archivo
    io.Copy(w, file)
}

// GetStats obtiene estadísticas
func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
    stats, err := s.db.GetStats()
    if err != nil {
        http.Error(w, "Error fetching stats", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

// saveAndHash guarda archivo y calcula hashes simultáneamente
func (s *Server) saveAndHash(src io.Reader, dest io.Writer) (blake3Hash, sha256Hash string, size int64, err error) {
    // TeeReader para calcular hashes mientras copiamos
    b3, s256, tee := hash.NewDualHasher(src)
    
    written, err := io.Copy(dest, tee)
    if err != nil {
        return "", "", 0, err
    }
    
    return b3.Finish(), s256.Finish(), written, nil
}
```

#### Routes & Server (6-8h)

**`internal/api/server.go`**

```go
package api

import (
    "net/http"
    
    "github.com/yourusername/fccur/internal/storage"
)

type Server struct {
    db          *storage.Database
    packagesDir string
    mux         *http.ServeMux
}

func NewServer(db *storage.Database, packagesDir string) *Server {
    s := &Server{
        db:          db,
        packagesDir: packagesDir,
        mux:         http.NewServeMux(),
    }
    
    s.setupRoutes()
    return s
}

func (s *Server) setupRoutes() {
    // API routes
    s.mux.HandleFunc("/api/packages", s.withCORS(s.withLogging(s.GetPackages)))
    s.mux.HandleFunc("/api/packages/", s.withCORS(s.withLogging(s.GetPackage)))
    s.mux.HandleFunc("/api/upload", s.withCORS(s.withLogging(s.UploadPackage)))
    s.mux.HandleFunc("/download/", s.withCORS(s.withLogging(s.DownloadPackage)))
    s.mux.HandleFunc("/api/stats", s.withCORS(s.withLogging(s.GetStats)))
    s.mux.HandleFunc("/health", s.Health)
    
    // Static files
    s.mux.Handle("/", http.FileServer(http.Dir("./web")))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status":"ok"}`))
}
```

**`internal/api/middleware.go`**

```go
package api

import (
    "log"
    "net/http"
    "time"
)

func (s *Server) withCORS(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next(w, r)
    }
}

func (s *Server) withLogging(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        next(w, r)
        
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    }
}
```

#### Checklist Día 2

- [ ] Database operations completas (CRUD)
- [ ] Todos los endpoints API implementados
- [ ] File upload funcional con dual hashing
- [ ] File download con streaming
- [ ] Statistics tracking
- [ ] CORS configurado
- [ ] Logging básico
- [ ] Tests manuales con curl/Postman

**Output esperado**: 
```bash
# Upload
curl -X POST -F "package=@ubuntu.iso" -F "name=Ubuntu" \
  -F "version=22.04" -F "category=os" http://localhost:8080/api/upload

# List
curl http://localhost:8080/api/packages

# Download
curl -o test.iso http://localhost:8080/download/1
```

---

### **DÍA 3: Frontend & UI** ⏱️ 8 horas

**Objetivo**: Interfaz web funcional y responsive

#### HTML Structure (0-2h)

**`web/index.html`**

```html
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FCCUR - Free Community Content Universal Repository</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <header>
        <div class="container">
            <h1>📦 FCCUR</h1>
            <p>Free Community Content Universal Repository</p>
        </div>
    </header>

    <main class="container">
        <!-- Search & Filter -->
        <section id="search-section">
            <input type="text" id="search-input" placeholder="Buscar paquetes...">
            <select id="category-filter">
                <option value="">Todas las categorías</option>
                <option value="os">Sistemas Operativos</option>
                <option value="compiler">Compiladores</option>
                <option value="ide">IDEs</option>
                <option value="tool">Herramientas</option>
                <option value="library">Bibliotecas</option>
            </select>
        </section>

        <!-- Admin Upload (oculto por defecto) -->
        <section id="admin-section" style="display: none;">
            <h2>Subir Paquete</h2>
            <form id="upload-form">
                <input type="text" name="name" placeholder="Nombre" required>
                <input type="text" name="version" placeholder="Versión" required>
                <select name="category" required>
                    <option value="">Seleccionar categoría</option>
                    <option value="os">Sistema Operativo</option>
                    <option value="compiler">Compilador</option>
                    <option value="ide">IDE</option>
                    <option value="tool">Herramienta</option>
                    <option value="library">Biblioteca</option>
                </select>
                <select name="platform" required>
                    <option value="">Plataforma</option>
                    <option value="windows">Windows</option>
                    <option value="mac">macOS</option>
                    <option value="linux">Linux</option>
                    <option value="all">Todas</option>
                </select>
                <textarea name="description" placeholder="Descripción"></textarea>
                <input type="file" name="package" required>
                <button type="submit">Subir Paquete</button>
            </form>
            <div id="upload-progress" style="display: none;">
                <div class="progress-bar">
                    <div class="progress-fill"></div>
                </div>
                <p class="progress-text">0%</p>
            </div>
        </section>

        <!-- Packages Grid -->
        <section id="packages-section">
            <div id="packages-grid"></div>
            <div id="loading">Cargando paquetes...</div>
            <div id="error" style="display: none;"></div>
        </section>

        <!-- Stats -->
        <section id="stats-section">
            <h2>Estadísticas</h2>
            <div id="stats-grid"></div>
        </section>
    </main>

    <footer>
        <p>&copy; 2025 FCCUR - Sistema Local de Distribución Educativa</p>
        <p>
            <button id="admin-toggle">👨‍💼 Admin</button>
            <a href="https://github.com/yourusername/fccur" target="_blank">GitHub</a>
        </p>
    </footer>

    <script src="app.js"></script>
</body>
</html>
```

#### CSS Styling (2-4h)

**`web/style.css`**

```css
/* Reset & Base */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

:root {
    --primary: #2563eb;
    --primary-dark: #1d4ed8;
    --success: #10b981;
    --danger: #ef4444;
    --gray-50: #f9fafb;
    --gray-100: #f3f4f6;
    --gray-200: #e5e7eb;
    --gray-300: #d1d5db;
    --gray-700: #374151;
    --gray-900: #111827;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    line-height: 1.6;
    color: var(--gray-900);
    background: var(--gray-50);
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1rem;
}

/* Header */
header {
    background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
    color: white;
    padding: 2rem 0;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

header h1 {
    font-size: 2rem;
    margin-bottom: 0.5rem;
}

header p {
    opacity: 0.9;
}

/* Search Section */
#search-section {
    display: flex;
    gap: 1rem;
    margin: 2rem 0;
    flex-wrap: wrap;
}

#search-input,
#category-filter {
    flex: 1;
    min-width: 200px;
    padding: 0.75rem 1rem;
    border: 2px solid var(--gray-200);
    border-radius: 8px;
    font-size: 1rem;
}

#search-input:focus,
#category-filter:focus {
    outline: none;
    border-color: var(--primary);
}

/* Admin Section */
#admin-section {
    background: white;
    padding: 2rem;
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    margin: 2rem 0;
}

#upload-form {
    display: grid;
    gap: 1rem;
}

#upload-form input,
#upload-form select,
#upload-form textarea {
    padding: 0.75rem;
    border: 2px solid var(--gray-200);
    border-radius: 8px;
    font-size: 1rem;
}

#upload-form textarea {
    min-height: 100px;
    resize: vertical;
}

#upload-form button {
    background: var(--primary);
    color: white;
    padding: 1rem;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
}

#upload-form button:hover {
    background: var(--primary-dark);
}

.progress-bar {
    width: 100%;
    height: 8px;
    background: var(--gray-200);
    border-radius: 4px;
    overflow: hidden;
    margin: 1rem 0;
}

.progress-fill {
    height: 100%;
    background: var(--success);
    transition: width 0.3s;
    width: 0%;
}

.progress-text {
    text-align: center;
    color: var(--gray-700);
}

/* Packages Grid */
#packages-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
    margin: 2rem 0;
}

.package-card {
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    transition: transform 0.2s, box-shadow 0.2s;
}

.package-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 4px 16px rgba(0,0,0,0.15);
}

.package-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 1rem;
}

.package-header h3 {
    font-size: 1.25rem;
    color: var(--gray-900);
}

.version {
    background: var(--gray-100);
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.875rem;
    color: var(--gray-700);
}

.package-body {
    margin-bottom: 1rem;
}

.description {
    color: var(--gray-700);
    font-size: 0.875rem;
    margin-bottom: 1rem;
}

.package-meta {
    display: flex;
    gap: 1rem;
    font-size: 0.875rem;
    color: var(--gray-700);
    margin-bottom: 1rem;
}

.package-footer {
    display: flex;
    gap: 0.5rem;
}

.btn-download,
.btn-info {
    flex: 1;
    padding: 0.75rem;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-download {
    background: var(--primary);
    color: white;
}

.btn-download:hover {
    background: var(--primary-dark);
}

.btn-info {
    background: var(--gray-100);
    color: var(--gray-700);
}

.btn-info:hover {
    background: var(--gray-200);
}

/* Stats */
#stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
    margin: 1rem 0;
}

.stat-card {
    background: white;
    padding: 1.5rem;
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.stat-card h4 {
    color: var(--gray-700);
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
}

.stat-card .value {
    font-size: 2rem;
    font-weight: bold;
    color: var(--primary);
}

/* Loading & Error */
#loading,
#error {
    text-align: center;
    padding: 3rem;
    color: var(--gray-700);
}

#error {
    color: var(--danger);
}

/* Footer */
footer {
    background: var(--gray-900);
    color: white;
    text-align: center;
    padding: 2rem;
    margin-top: 4rem;
}

footer p {
    margin: 0.5rem 0;
}

footer a,
footer button {
    color: var(--primary);
    text-decoration: none;
    background: none;
    border: none;
    cursor: pointer;
    font-size: 1rem;
}

footer a:hover,
footer button:hover {
    text-decoration: underline;
}

/* Modal */
.modal {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0,0,0,0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.modal-content {
    background: white;
    padding: 2rem;
    border-radius: 12px;
    max-width: 600px;
    width: 90%;
    max-height: 80vh;
    overflow-y: auto;
    position: relative;
}

.close {
    position: absolute;
    top: 1rem;
    right: 1rem;
    font-size: 2rem;
    cursor: pointer;
    color: var(--gray-700);
}

.close:hover {
    color: var(--gray-900);
}

.info-grid {
    display: grid;
    gap: 1rem;
    margin: 1rem 0;
}

.info-item {
    padding: 1rem;
    background: var(--gray-50);
    border-radius: 8px;
}

.info-item strong {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--gray-900);
}

.hash {
    font-family: 'Courier New', monospace;
    font-size: 0.75rem;
    word-break: break-all;
    color: var(--gray-700);
    background: white;
    padding: 0.5rem;
    border-radius: 4px;
}

/* Responsive */
@media (max-width: 768px) {
    header h1 {
        font-size: 1.5rem;
    }
    
    #packages-grid {
        grid-template-columns: 1fr;
    }
    
    #search-section {
        flex-direction: column;
    }
}
```

#### JavaScript Logic (4-8h)

**`web/app.js`**

```javascript
// API Base URL
const API_BASE = '/api';

// Estado global
let allPackages = [];
let filteredPackages = [];

// Inicializar
document.addEventListener('DOMContentLoaded', () => {
    loadPackages();
    loadStats();
    setupEventListeners();
});

// Event Listeners
function setupEventListeners() {
    // Search
    document.getElementById('search-input').addEventListener('input', (e) => {
        filterPackages(e.target.value, null);
    });
    
    // Category filter
    document.getElementById('category-filter').addEventListener('change', (e) => {
        filterPackages(null, e.target.value);
    });
    
    // Admin toggle
    document.getElementById('admin-toggle').addEventListener('click', toggleAdmin);
    
    // Upload form
    document.getElementById('upload-form').addEventListener('submit', uploadPackage);
}

// Cargar paquetes
async function loadPackages() {
    try {
        showLoading();
        const response = await fetch(`${API_BASE}/packages`);
        
        if (!response.ok) {
            throw new Error('Error loading packages');
        }
        
        allPackages = await response.json() || [];
        filteredPackages = [...allPackages];
        renderPackages();
        hideLoading();
    } catch (error) {
        showError(error.message);
        hideLoading();
    }
}

// Renderizar paquetes
function renderPackages() {
    const grid = document.getElementById('packages-grid');
    grid.innerHTML = '';
    
    if (filteredPackages.length === 0) {
        grid.innerHTML = '<p class="no-results">No se encontraron paquetes</p>';
        return;
    }
    
    filteredPackages.forEach(pkg => {
        const card = createPackageCard(pkg);
        grid.appendChild(card);
    });
}

// Crear card de paquete
function createPackageCard(pkg) {
    const card = document.createElement('div');
    card.className = 'package-card';
    
    const size = formatFileSize(pkg.file_size);
    const platform = getPlatformIcon(pkg.platform);
    
    card.innerHTML = `
        <div class="package-header">
            <h3>${pkg.name}</h3>
            <span class="version">v${pkg.version}</span>
        </div>
        <div class="package-body">
            <p class="description">${pkg.description || 'Sin descripción'}</p>
            <div class="package-meta">
                <span class="platform">${platform} ${pkg.platform}</span>
                <span class="size">📦 ${size}</span>
            </div>
        </div>
        <div class="package-footer">
            <button class="btn-download" onclick="downloadPackage(${pkg.id})">
                ⬇️ Descargar
            </button>
            <button class="btn-info" onclick="showPackageInfo(${pkg.id})">
                ℹ️ Info
            </button>
        </div>
    `;
    
    return card;
}

// Descargar paquete
async function downloadPackage(id) {
    try {
        const pkg = allPackages.find(p => p.id === id);
        if (!pkg) return;
        
        // Crear elemento <a> para download
        const a = document.createElement('a');
        a.href = `/download/?id=${id}`;
        a.download = `${pkg.name}-${pkg.version}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        
        // Mostrar notificación
        showSuccess('Descarga iniciada');
        
        // Recargar stats
        setTimeout(loadStats, 1000);
    } catch (error) {
        showError('Error al descargar: ' + error.message);
    }
}

// Mostrar info del paquete
function showPackageInfo(id) {
    const pkg = allPackages.find(p => p.id === id);
    if (!pkg) return;
    
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <span class="close" onclick="this.parentElement.parentElement.remove()">&times;</span>
            <h2>${pkg.name} v${pkg.version}</h2>
            <div class="info-grid">
                <div class="info-item">
                    <strong>Categoría:</strong> ${pkg.category}
                </div>
                <div class="info-item">
                    <strong>Plataforma:</strong> ${pkg.platform}
                </div>
                <div class="info-item">
                    <strong>Tamaño:</strong> ${formatFileSize(pkg.file_size)}
                </div>
                <div class="info-item">
                    <strong>BLAKE3 Hash:</strong>
                    <code class="hash">${pkg.blake3_hash}</code>
                </div>
                <div class="info-item">
                    <strong>SHA256 Hash:</strong>
                    <code class="hash">${pkg.sha256_hash}</code>
                </div>
                ${pkg.download_url ? `
                <div class="info-item">
                    <strong>URL Original:</strong>
                    <a href="${pkg.download_url}" target="_blank">${pkg.download_url}</a>
                </div>
                ` : ''}
            </div>
            <p class="description-full">${pkg.description || 'Sin descripción'}</p>
        </div>
    `;
    
    document.body.appendChild(modal);
}

// Filtrar paquetes
function filterPackages(searchTerm, category) {
    const search = searchTerm?.toLowerCase() || document.getElementById('search-input').value.toLowerCase();
    const cat = category || document.getElementById('category-filter').value;
    
    filteredPackages = allPackages.filter(pkg => {
        const matchesSearch = !search ||
            pkg.name.toLowerCase().includes(search) ||
            pkg.description?.toLowerCase().includes(search);
        
        const matchesCategory = !cat || pkg.category === cat;
        
        return matchesSearch && matchesCategory;
    });
    
    renderPackages();
}

// Upload paquete
async function uploadPackage(e) {
    e.preventDefault();
    
    const form = e.target;
    const formData = new FormData(form);
    
    try {
        showUploadProgress();
        
        const response = await fetch(`${API_BASE}/upload`, {
            method: 'POST',
            body: formData
        });
        
        if (!response.ok) {
            throw new Error('Error uploading package');
        }
        
        const pkg = await response.json();
        
        hideUploadProgress();
        showSuccess('Paquete subido exitosamente');
        form.reset();
        
        // Recargar paquetes
        await loadPackages();
    } catch (error) {
        hideUploadProgress();
        showError('Error al subir paquete: ' + error.message);
    }
}

// Stats
async function loadStats() {
    try {
        const response = await fetch(`${API_BASE}/stats`);
        if (!response.ok) return;
        
        const stats = await response.json();
        renderStats(stats);
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

function renderStats(stats) {
    const grid = document.getElementById('stats-grid');
    grid.innerHTML = '';
    
    // Total packages
    const totalCard = document.createElement('div');
    totalCard.className = 'stat-card';
    totalCard.innerHTML = `
        <h4>Total Paquetes</h4>
        <div class="value">${allPackages.length}</div>
    `;
    grid.appendChild(totalCard);
    
    // Total downloads
    const totalDownloads = stats.reduce((sum, s) => sum + s.total_downloads, 0);
    const downloadsCard = document.createElement('div');
    downloadsCard.className = 'stat-card';
    downloadsCard.innerHTML = `
        <h4>Total Descargas</h4>
        <div class="value">${totalDownloads}</div>
    `;
    grid.appendChild(downloadsCard);
    
    // Most downloaded
    if (stats.length > 0) {
        const topPkg = stats[0];
        const topCard = document.createElement('div');
        topCard.className = 'stat-card';
        topCard.innerHTML = `
            <h4>Más Descargado</h4>
            <div class="value">${topPkg.package_name}</div>
            <p style="color: var(--gray-700); font-size: 0.875rem; margin-top: 0.5rem;">
                ${topPkg.total_downloads} descargas
            </p>
        `;
        grid.appendChild(topCard);
    }
}

// Utilidades
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

function getPlatformIcon(platform) {
    const icons = {
        'windows': '🪟',
        'mac': '🍎',
        'linux': '🐧',
        'all': '💻'
    };
    return icons[platform.toLowerCase()] || '💻';
}

function toggleAdmin() {
    const section = document.getElementById('admin-section');
    section.style.display = section.style.display === 'none' ? 'block' : 'none';
}

function showLoading() {
    document.getElementById('loading').style.display = 'block';
    document.getElementById('packages-grid').style.display = 'none';
}

function hideLoading() {
    document.getElementById('loading').style.display = 'none';
    document.getElementById('packages-grid').style.display = 'grid';
}

function showError(message) {
    const errorDiv = document.getElementById('error');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}

function showSuccess(message) {
    // Simple alert por ahora
    alert(message);
}

function showUploadProgress() {
    document.getElementById('upload-progress').style.display = 'block';
}

function hideUploadProgress() {
    document.getElementById('upload-progress').style.display = 'none';
}
```

#### Checklist Día 3

- [ ] HTML structure completa
- [ ] CSS responsive implementado
- [ ] JavaScript Fetch API funcional
- [ ] Package listing con search/filter
- [ ] Upload form funcional
- [ ] Download con progress indicator
- [ ] Package info modal
- [ ] Stats dashboard
- [ ] Testeo en mobile y desktop

---

### **DÍA 4: Deployment & Testing** ⏱️ 8 horas

**Objetivo**: Sistema deployable en Raspberry Pi

#### Build & Deployment Scripts (0-3h)

**`Makefile`**

```makefile
.PHONY: build build-pi run test clean deploy

# Variables
BINARY_NAME=fccur
PI_USER=pi
PI_HOST=192.168.1.100
PI_DIR=/home/pi/fccur

build:
	go build -o $(BINARY_NAME) cmd/server/main.go

build-pi:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-arm64 cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-arm64
	rm -f data/*.db

deploy: build-pi
	@echo "📦 Deploying to Raspberry Pi..."
	ssh $(PI_USER)@$(PI_HOST) "mkdir -p $(PI_DIR)/{packages,data,web}"
	scp $(BINARY_NAME)-arm64 $(PI_USER)@$(PI_HOST):$(PI_DIR)/$(BINARY_NAME)
	scp -r web/* $(PI_USER)@$(PI_HOST):$(PI_DIR)/web/
	scp deploy/fccur.service $(PI_USER)@$(PI_HOST):~/
	ssh $(PI_USER)@$(PI_HOST) "sudo mv ~/fccur.service /etc/systemd/system/"
	ssh $(PI_USER)@$(PI_HOST) "sudo systemctl daemon-reload"
	ssh $(PI_USER)@$(PI_HOST) "sudo systemctl enable fccur"
	ssh $(PI_USER)@$(PI_HOST) "sudo systemctl restart fccur"
	@echo "✅ Deployment complete!"
```

**`deploy/fccur.service`**

```ini
[Unit]
Description=FCCUR - Free Community Content Universal Repository
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/fccur
ExecStart=/home/pi/fccur/fccur -addr=:8080 -db=/home/pi/fccur/data/fccur.db -packages=/home/pi/fccur/packages
Restart=on-failure
RestartSec=5s

# Resource limits
MemoryLimit=512M
CPUQuota=50%

# Logging
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**`deploy/setup-pi.sh`**

```bash
#!/bin/bash
# Setup script para Raspberry Pi

set -e

echo "🔧 Setting up FCCUR on Raspberry Pi..."

# Update system
sudo apt update
sudo apt upgrade -y

# Install dependencies
sudo apt install -y sqlite3 ufw

# Configure firewall
sudo ufw allow 22/tcp  # SSH
sudo ufw allow 8080/tcp  # FCCUR
sudo ufw --force enable

# Create directories
mkdir -p ~/fccur/{packages,data,web}

# Configure static IP (optional)
read -p "Configure static IP? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Edit /etc/dhcpcd.conf and add:"
    echo "interface eth0"
    echo "static ip_address=192.168.1.100/24"
    echo "static routers=192.168.1.1"
    echo "static domain_name_servers=8.8.8.8 8.8.4.4"
fi

echo "✅ Setup complete! Now run 'make deploy' from your dev machine."
```

#### Integration Tests (3-6h)

**`tests/integration_test.go`**

```go
package tests

import (
    "bytes"
    "encoding/json"
    "io"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/yourusername/fccur/internal/api"
    "github.com/yourusername/fccur/internal/models"
    "github.com/yourusername/fccur/internal/storage"
)

func setupTestServer(t *testing.T) (*api.Server, func()) {
    // Create temp database
    dbPath := filepath.Join(t.TempDir(), "test.db")
    db, err := storage.NewDatabase(dbPath)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    
    if err := db.Migrate(); err != nil {
        t.Fatalf("Failed to migrate: %v", err)
    }
    
    // Create temp packages dir
    packagesDir := t.TempDir()
    
    server := api.NewServer(db, packagesDir)
    
    cleanup := func() {
        db.Close()
    }
    
    return server, cleanup
}

func TestHealthEndpoint(t *testing.T) {
    server, cleanup := setupTestServer(t)
    defer cleanup()
    
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()
    
    server.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    var response map[string]string
    json.NewDecoder(w.Body).Decode(&response)
    
    if response["status"] != "ok" {
        t.Errorf("Expected status ok, got %s", response["status"])
    }
}

func TestUploadAndDownload(t *testing.T) {
    server, cleanup := setupTestServer(t)
    defer cleanup()
    
    // Create test file
    testContent := []byte("test package content")
    
    // Upload
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    part, _ := writer.CreateFormFile("package", "test.txt")
    part.Write(testContent)
    
    writer.WriteField("name", "Test Package")
    writer.WriteField("version", "1.0.0")
    writer.WriteField("category", "test")
    writer.WriteField("platform", "all")
    writer.WriteField("description", "Test description")
    
    writer.Close()
    
    req := httptest.NewRequest("POST", "/api/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    w := httptest.NewRecorder()
    
    server.ServeHTTP(w, req)
    
    if w.Code != http.StatusCreated {
        t.Fatalf("Upload failed: %d - %s", w.Code, w.Body.String())
    }
    
    var pkg models.Package
    json.NewDecoder(w.Body).Decode(&pkg)
    
    // Download
    req = httptest.NewRequest("GET", "/download/?id="+string(pkg.ID), nil)
    w = httptest.NewRecorder()
    
    server.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Fatalf("Download failed: %d", w.Code)
    }
    
    downloaded, _ := io.ReadAll(w.Body)
    if !bytes.Equal(downloaded, testContent) {
        t.Errorf("Downloaded content doesn't match")
    }
}

func TestListPackages(t *testing.T) {
    server, cleanup := setupTestServer(t)
    defer cleanup()
    
    req := httptest.NewRequest("GET", "/api/packages", nil)
    w := httptest.NewRecorder()
    
    server.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    var packages []*models.Package
    json.NewDecoder(w.Body).Decode(&packages)
    
    // Initially should be empty
    if len(packages) != 0 {
        t.Errorf("Expected 0 packages, got %d", len(packages))
    }
}
```

#### Monitoring & Logging (6-8h)

**Health Check Endpoint**

```go
// internal/api/handlers.go

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
    // Check database
    if err := s.db.Ping(); err != nil {
        http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
        return
    }
    
    // Check packages directory
    if _, err := os.Stat(s.packagesDir); os.IsNotExist(err) {
        http.Error(w, "Packages directory missing", http.StatusServiceUnavailable)
        return
    }
    
    // Get stats
    stats, _ := s.db.GetStats()
    totalDownloads := 0
    for _, s := range stats {
        totalDownloads += s.TotalDownloads
    }
    
    response := map[string]interface{}{
        "status": "ok",
        "total_packages": len(stats),
        "total_downloads": totalDownloads,
        "uptime": time.Since(s.startTime).String(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

**Structured Logging**

```go
// internal/api/middleware.go

type responseWriter struct {
    http.ResponseWriter
    statusCode int
    bytes      int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    n, err := rw.ResponseWriter.Write(b)
    rw.bytes += n
    return n, err
}

func (s *Server) withLogging(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next(rw, r)
        
        duration := time.Since(start)
        
        log.Printf("[%s] %s %s %d %d %v %s",
            start.Format("2006-01-02 15:04:05"),
            r.Method,
            r.URL.Path,
            rw.statusCode,
            rw.bytes,
            duration,
            r.RemoteAddr,
        )
    }
}
```

#### Checklist Día 4

- [ ] Makefile con comandos útiles
- [ ] Cross-compilation para ARM64
- [ ] Systemd service file
- [ ] Setup script para Pi
- [ ] Integration tests pasando
- [ ] Health check endpoint
- [ ] Structured logging
- [ ] Deployed en Raspberry Pi
- [ ] Accessible en red local
- [ ] 3+ personas lo probaron exitosamente

---

### **DÍA 5: Polish & Documentation** ⏱️ 8 horas

**Objetivo**: Sistema production-ready con documentación completa

#### Documentation (0-4h)

**README.md completo** (este archivo que estás leyendo)

**`docs/API.md`**

```markdown
# FCCUR API Documentation

## Base URL

```
http://localhost:8080
```

## Endpoints

### Health Check

```
GET /health
```

**Response**:
```json
{
  "status": "ok",
  "total_packages": 12,
  "total_downloads": 247,
  "uptime": "2h34m12s"
}
```

### List Packages

```
GET /api/packages
```

**Response**:
```json
[
  {
    "id": 1,
    "name": "Ubuntu 22.04 LTS",
    "version": "22.04.3",
    "description": "Ubuntu Desktop ISO",
    "category": "os",
    "file_path": "/packages/ubuntu-22.04.iso",
    "file_size": 4700000000,
    "blake3_hash": "abc123...",
    "sha256_hash": "def456...",
    "platform": "linux",
    "created_at": "2025-01-01T10:00:00Z",
    "updated_at": "2025-01-01T10:00:00Z"
  }
]
```

### Get Package

```
GET /api/packages/?id=1
```

### Upload Package

```
POST /api/upload
Content-Type: multipart/form-data
```

**Form Fields**:
- `package` (file): Package file
- `name` (string): Package name
- `version` (string): Version
- `category` (string): Category (os, compiler, ide, tool, library)
- `platform` (string): Platform (windows, mac, linux, all)
- `description` (string, optional): Description

### Download Package

```
GET /download/?id=1
```

**Response Headers**:
- `Content-Disposition`: attachment; filename="..."
- `Content-Length`: File size
- `X-BLAKE3-Hash`: BLAKE3 hash
- `X-SHA256-Hash`: SHA256 hash

### Statistics

```
GET /api/stats
```

**Response**:
```json
[
  {
    "package_id": 1,
    "package_name": "Ubuntu 22.04 LTS",
    "total_downloads": 45,
    "last_download": "2025-01-15T14:30:00Z"
  }
]
```
```

**`docs/DEPLOYMENT.md`**

```markdown
# Deployment Guide

## Prerequisites

- Raspberry Pi 4 (4GB RAM recommended)
- USB 3.0 SSD (256GB+)
- Ethernet cable (Gigabit recommended)
- Raspberry Pi OS (64-bit)

## Step 1: Prepare Raspberry Pi

```bash
# SSH into your Pi
ssh pi@192.168.1.100

# Run setup script
curl -sSL https://raw.githubusercontent.com/yourusername/fccur/main/deploy/setup-pi.sh | bash
```

## Step 2: Deploy from Dev Machine

```bash
# Build and deploy
make deploy

# Or manually:
GOOS=linux GOARCH=arm64 go build -o fccur-arm64 cmd/server/main.go
scp fccur-arm64 pi@192.168.1.100:~/fccur/fccur
scp -r web/* pi@192.168.1.100:~/fccur/web/
```

## Step 3: Configure Service

```bash
# On Pi
sudo systemctl enable fccur
sudo systemctl start fccur
sudo systemctl status fccur
```

## Step 4: Verify

```bash
# Check health
curl http://192.168.1.100:8080/health

# Check web interface
# Open browser: http://192.168.1.100:8080
```

## Troubleshooting

### Service won't start

```bash
# Check logs
sudo journalctl -u fccur -n 50 -f

# Check permissions
ls -la ~/fccur/

# Restart service
sudo systemctl restart fccur
```

### Can't access from network

```bash
# Check firewall
sudo ufw status

# Allow port
sudo ufw allow 8080/tcp

# Check if service is listening
sudo netstat -tulpn | grep 8080
```

### Database issues

```bash
# Verify database
sqlite3 ~/fccur/data/fccur.db "PRAGMA integrity_check;"

# Backup database
cp ~/fccur/data/fccur.db ~/fccur/data/fccur.db.backup
```
```

#### Package Pre-loading (4-6h)

**Script para cargar paquetes iniciales**

**`scripts/load-initial-packages.sh`**

```bash
#!/bin/bash
# Script para cargar paquetes iniciales

API_URL="http://localhost:8080/api/upload"

# Función para subir paquete
upload_package() {
    local file=$1
    local name=$2
    local version=$3
    local category=$4
    local platform=$5
    local description=$6
    
    echo "📦 Uploading $name $version..."
    
    curl -X POST "$API_URL" \
        -F "package=@$file" \
        -F "name=$name" \
        -F "version=$version" \
        -F "category=$category" \
        -F "platform=$platform" \
        -F "description=$description"
    
    echo ""
}

# Descargar y subir paquetes esenciales

# 1. VS Code
if [ ! -f "vscode.deb" ]; then
    echo "Downloading VS Code..."
    wget https://code.visualstudio.com/sha/download?build=stable&os=linux-deb-x64 -O vscode.deb
fi
upload_package "vscode.deb" "Visual Studio Code" "1.85" "ide" "linux" "Editor de código moderno y extensible"

# 2. Git
if [ ! -f "git.deb" ]; then
    apt download git
fi
upload_package "git*.deb" "Git" "2.43" "tool" "linux" "Sistema de control de versiones distribuido"

# 3. Python
echo "Subiendo Python..."
# (Asumiendo que ya tienes los instaladores)

# ... más paquetes

echo "✅ All packages uploaded!"
```

#### Final Testing & Bug Fixes (6-8h)

**Test Checklist**:

```markdown
## Functional Tests

- [ ] Health endpoint responds
- [ ] Can list packages
- [ ] Can upload package (< 100MB)
- [ ] Can upload large package (> 1GB)
- [ ] Can download package
- [ ] Download hash verification works
- [ ] Search/filter works
- [ ] Stats update after download
- [ ] Mobile UI works
- [ ] Offline mode works (PWA)

## Performance Tests

- [ ] 5 concurrent downloads work
- [ ] Upload doesn't block downloads
- [ ] Database queries < 100ms
- [ ] File streaming efficient (no memory spike)
- [ ] Server uses < 256MB RAM under load

## Security Tests

- [ ] Large file upload rejected (> 10GB)
- [ ] SQL injection attempts fail
- [ ] Path traversal attempts fail
- [ ] CORS works correctly
- [ ] Rate limiting works

## Deployment Tests

- [ ] Systemd service starts on boot
- [ ] Service restarts on failure
- [ ] Logs are accessible
- [ ] Firewall configured correctly
- [ ] Static IP works
```

#### Checklist Día 5

- [ ] README.md completo
- [ ] API documentation
- [ ] Deployment guide
- [ ] 10-12 packages pre-loaded
- [ ] All tests passing
- [ ] Known bugs documented
- [ ] Demo video/screenshots
- [ ] GitHub repo public
- [ ] 5+ external users tested
- [ ] Feedback collected

---

## 📥 Instalación y Setup

### Desarrollo Local

```bash
# 1. Clonar repositorio
git clone https://github.com/yourusername/fccur.git
cd fccur

# 2. Instalar dependencias
go mod download

# 3. Crear directorios
mkdir -p packages data web

# 4. Ejecutar servidor
go run cmd/server/main.go

# 5. Abrir navegador
open http://localhost:8080
```

### Producción (Raspberry Pi)

```bash
# En tu Pi
curl -sSL https://raw.githubusercontent.com/yourusername/fccur/main/deploy/setup-pi.sh | bash

# En tu máquina de desarrollo
make deploy
```

---

## 🎮 Uso

### Para Estudiantes

1. Abrir navegador: `http://fccur.local:8080` o `http://192.168.1.100:8080`
2. Buscar paquete deseado
3. Click en "Descargar"
4. Verificar hash (opcional):
   ```bash
   # Linux/Mac
   sha256sum ubuntu-22.04.iso
   ```

### Para Profesores

1. Click en botón "Admin" en el footer
2. Completar formulario de upload
3. Seleccionar archivo
4. Subir paquete
5. Verificar en lista de paquetes

---

## 📊 API Reference

Ver [docs/API.md](docs/API.md) para documentación completa.

**Quick Examples**:

```bash
# Listar todos los paquetes
curl http://localhost:8080/api/packages

# Descargar paquete
curl -o package.iso http://localhost:8080/download/?id=1

# Ver estadísticas
curl http://localhost:8080/api/stats

# Health check
curl http://localhost:8080/health
```

---

## 🚀 Deployment

### Opción 1: Systemd (Recomendado)

```bash
# Copiar service file
sudo cp deploy/fccur.service /etc/systemd/system/

# Habilitar y arrancar
sudo systemctl enable fccur
sudo systemctl start fccur

# Ver logs
sudo journalctl -u fccur -f
```

### Opción 2: Manual

```bash
# Compilar
go build -o fccur cmd/server/main.go

# Ejecutar
./fccur -addr=:8080 -db=./data/fccur.db -packages=./packages
```

### Opción 3: Docker (Futuro)

```bash
# TODO: Crear Dockerfile
docker build -t fccur .
docker run -p 8080:8080 -v $(pwd)/packages:/packages fccur
```

---

## 🗺️ Roadmap

### v0.1 (MVP) - Semana 1 ✅
- [x] Backend Go funcional
- [x] SQLite database
- [x] API REST completa
- [x] Web UI básica
- [x] Dual hashing (BLAKE3 + SHA256)
- [x] File upload/download
- [x] Basic stats

### v0.2 - Semana 2
- [ ] User authentication (Basic Auth)
- [ ] Admin panel mejorado
- [ ] Package categories advanced
- [ ] Download resume support
- [ ] Bandwidth throttling

### v0.3 - Semana 3
- [ ] Multi-node sync (CRDT básico)
- [ ] Service discovery (mDNS)
- [ ] Load balancing (Power of Two)
- [ ] Health monitoring dashboard
- [ ] Automated backups

### v1.0 - Mes 2
- [ ] Native sandboxing (namespaces)
- [ ] Web of Trust authentication
- [ ] GraphQL API
- [ ] Progressive Web App complete
- [ ] Mobile app (React Native)

### v2.0 - Mes 3+
- [ ] Multi-campus support
- [ ] Advanced analytics
- [ ] Package signing
- [ ] Automated testing
- [ ] CI/CD pipeline

---

## 🤝 Contribuir

¡Las contribuciones son bienvenidas!

### Proceso

1. Fork el repositorio
2. Crear branch (`git checkout -b feature/amazing-feature`)
3. Commit cambios (`git commit -m 'Add amazing feature'`)
4. Push al branch (`git push origin feature/amazing-feature`)
5. Abrir Pull Request

### Guías

- Seguir las convenciones de Go
- Escribir tests para nuevo código
- Actualizar documentación
- Un feature por PR

### Áreas donde puedes ayudar

**Fácil**:
- Mejorar documentación
- Agregar más scripts de instalación
- Traducir interfaz
- Reportar bugs

**Medio**:
- Implementar nuevos endpoints
- Mejorar UI/UX
- Optimizar queries
- Agregar validaciones

**Difícil**:
- Sistema de autenticación
- Multi-node sync
- Native sandboxing
- Advanced analytics

---

## 📝 Licencia

MIT License - ver [LICENSE](LICENSE) para detalles.

---

## 🙏 Agradecimientos

- Dra. Hilda Castillo Zacatelco - Mentora del proyecto
- Estudiantes de FCC BUAP - Beta testers
- Comunidad Open Source - Por las herramientas increíbles

---

## 📧 Contacto

- **Proyecto**: [github.com/yourusername/fccur](https://github.com/yourusername/fccur)
- **Issues**: [github.com/yourusername/fccur/issues](https://github.com/yourusername/fccur/issues)
- **Email**: your.email@example.com

---

## 🎯 Filosofía del Proyecto

> "Simple, but not simplistic" - Albert Einstein

FCCUR abraza la simplicidad pragmática:

- ✅ **Tecnología simple y probada** (no microservicios complejos)
- ✅ **Un solo binario** (fácil de deployar)
- ✅ **SQLite** (sin administración de DB)
- ✅ **Funciona offline** (no requiere internet)
- ✅ **Código legible** (cualquier estudiante puede contribuir)

### No hacemos

- ❌ Over-engineering
- ❌ Frameworks complejos innecesarios
- ❌ Arquitecturas que requieren PhD para entender
- ❌ Features que usarías 1 vez al año

### Sí hacemos

- ✅ Resolver problemas reales
- ✅ Código mantenible
- ✅ Performance excelente
- ✅ Experiencia de usuario simple
- ✅ Documentación clara

---

**¡Empecemos a construir! 🚀**

Si tienes preguntas, abre un issue en GitHub. Si quieres contribuir, ¡adelante!

**Built with ❤️ for students, by students.**
