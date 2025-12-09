package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jesus/FCCUR/internal/storage"
)

func main() {
	var (
		dbPath         = flag.String("db", "./data/fccur.db", "Database connection string")
		migrationsPath = flag.String("migrations", "./migrations", "Path to migrations directory")
		command        = flag.String("command", "up", "Migration command: up, down, version, force, drop, goto, steps")
		version        = flag.Uint("version", 0, "Target version for 'goto' command")
		steps          = flag.Int("steps", 0, "Number of steps for 'steps' command (negative for down)")
		forceVersion   = flag.Int("force-version", -1, "Force version without running migrations")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "FCCUR Migration Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  up          Apply all pending migrations\n")
		fmt.Fprintf(os.Stderr, "  down        Rollback the most recent migration\n")
		fmt.Fprintf(os.Stderr, "  version     Show current migration version\n")
		fmt.Fprintf(os.Stderr, "  goto        Migrate to a specific version\n")
		fmt.Fprintf(os.Stderr, "  steps       Run n migrations (use negative for rollback)\n")
		fmt.Fprintf(os.Stderr, "  force       Force set version without running migrations\n")
		fmt.Fprintf(os.Stderr, "  drop        Drop all tables (DANGEROUS!)\n")
		fmt.Fprintf(os.Stderr, "  validate    Validate migration state\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -command=up\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=down\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=version\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=goto -version=2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=steps -steps=-2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=force -force-version=1\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -db=\"postgres://user:pass@localhost/db\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()

	// Initialize database
	db, err := storage.NewDatabase(*dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create migrator
	migrator, err := storage.NewMigrator(db, *migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer migrator.Close()

	dbType := migrator.GetDatabaseType()
	log.Printf("Database Type: %s", dbType)
	log.Printf("Migrations Path: %s/%s", *migrationsPath, dbType)

	// Execute command
	switch *command {
	case "up":
		log.Println("Running migrations...")
		if err := migrator.Up(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✓ Migrations applied successfully")

		version, dirty, _ := migrator.Version()
		log.Printf("Current version: %d (dirty: %v)", version, dirty)

	case "down":
		log.Println("Rolling back migration...")
		if err := migrator.Down(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("✓ Migration rolled back successfully")

		version, dirty, _ := migrator.Version()
		log.Printf("Current version: %d (dirty: %v)", version, dirty)

	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		log.Printf("Current version: %d", version)
		log.Printf("Dirty: %v", dirty)

		if dirty {
			log.Println("⚠ Database is in dirty state! Use 'force' command to fix.")
		}

	case "goto":
		if *version == 0 {
			log.Fatal("Please specify -version flag")
		}
		log.Printf("Migrating to version %d...", *version)
		if err := migrator.Goto(*version); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Printf("✓ Migrated to version %d successfully", *version)

	case "steps":
		if *steps == 0 {
			log.Fatal("Please specify -steps flag")
		}
		direction := "forward"
		if *steps < 0 {
			direction = "backward"
		}
		log.Printf("Running %d steps %s...", abs(*steps), direction)
		if err := migrator.Steps(*steps); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Printf("✓ Ran %d steps successfully", abs(*steps))

		version, dirty, _ := migrator.Version()
		log.Printf("Current version: %d (dirty: %v)", version, dirty)

	case "force":
		if *forceVersion < 0 {
			log.Fatal("Please specify -force-version flag")
		}
		log.Printf("⚠ Forcing version to %d without running migrations...", *forceVersion)
		if err := migrator.Force(*forceVersion); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		log.Printf("✓ Version forced to %d", *forceVersion)

	case "drop":
		log.Println("⚠ WARNING: This will drop all tables!")
		fmt.Print("Type 'yes' to continue: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			log.Println("Aborted")
			return
		}

		log.Println("Dropping all tables...")
		if err := migrator.Drop(); err != nil {
			log.Fatalf("Drop failed: %v", err)
		}
		log.Println("✓ All tables dropped")

	case "validate":
		log.Println("Validating migrations...")
		if err := migrator.Validate(); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
		log.Println("✓ Migrations are valid")

		version, dirty, _ := migrator.Version()
		log.Printf("Current version: %d (dirty: %v)", version, dirty)

	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
