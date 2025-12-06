# Supabase Integration Guide

## Why Use Supabase?

Supabase is a PostgreSQL-based database that can be used for:
- Storing metadata about sources
- Query history and analytics
- User management (if you want to add authentication)
- Source tracking and versioning

## Setup Supabase

### 1. Create Supabase Project

1. Go to https://supabase.com
2. Sign up / Login
3. Create a new project
4. Wait for the project to be ready

### 2. Get Connection String

In your Supabase project:
1. Go to **Settings** → **Database**
2. Find **Connection string** → **URI**
3. Copy the connection string

Format:
```
postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres
```

### 3. Configure .env

Add to your `.env` file:
```env
# Supabase Database
DB_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres

# Or using connection pooling (recommended for production)
DB_URL=postgresql://postgres.[PROJECT-REF]:[YOUR-PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres?pgbouncer=true
```

**Important:** Replace:
- `[YOUR-PASSWORD]` - Your database password
- `[PROJECT-REF]` - Your project reference ID
- `[REGION]` - Your project region (e.g., eu-central-1)

### 4. Enable Database in Application

Edit `cmd/main.go` to enable database connection:

```go
func init() {
    config.LoadEnvVariables()
    repositories.ConnectToDatabase() // Add this line
    migrations.RunMigrations(repositories.DB) // Add this if you have migrations
}
```

## Use Cases

### 1. Store Source Metadata

You can store information about sources added through UI:

```go
type Source struct {
    gorm.Model
    Title       string
    URL         string
    ContentHash string
    UserID      uint
    AddedAt     time.Time
}
```

### 2. Query History

Track all queries made:

```go
type QueryHistory struct {
    gorm.Model
    Question    string
    Answer      string
    Sources     []string
    UserID      uint
    CreatedAt   time.Time
}
```

### 3. User Management

If you want to add authentication:

```go
// Already exists in pkg/repositories/users.go
type User struct {
    gorm.Model
    Username  string
    Passwd    string
    FirstName string
    LastName  string
}
```

## Connection Pooling

For production, use Supabase's connection pooler:
- Port: **6543** (pooler) instead of 5432
- Add `?pgbouncer=true` to connection string
- Better performance and connection management

## Security

1. **Never commit `.env` file** - It's already in `.gitignore`
2. **Use environment variables** on server
3. **Enable Row Level Security (RLS)** in Supabase if needed
4. **Use connection pooling** for production

## Testing Connection

```go
// Test connection
if err := repositories.ConnectToDatabase(); err != nil {
    log.Fatal("Failed to connect to Supabase:", err)
}
log.Println("✅ Connected to Supabase!")
```

## Migration

If you have migrations, run them:

```go
migrations.RunMigrations(repositories.DB)
```

## Notes

- Supabase uses PostgreSQL, so all existing GORM code works
- Connection string format is the same as regular PostgreSQL
- Supabase provides a dashboard for managing data
- Free tier includes 500MB database storage

