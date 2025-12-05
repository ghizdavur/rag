# Deployment Guide - RAG Application

## Pregătire pentru Deployment

### 1. Push codul pe Git

```bash
git add .
git commit -m "Prepare for deployment"
git push origin main
```

### 2. Pe Server - Clonare Repository

```bash
# SSH în server
ssh user@your-server.com

# Clone repository
git clone https://github.com/ghizdavur/rag.git
cd rag
```

## Instalare Dependențe

### 3. Instalează Go

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Verifică instalarea
go version
```

### 4. Instalează Ollama

```bash
# Linux
curl -fsSL https://ollama.com/install.sh | sh

# Sau manual
# Descarcă de la: https://ollama.com/download/linux

# Pornește Ollama ca serviciu
sudo systemctl enable ollama
sudo systemctl start ollama

# Descarcă modelele
ollama pull nomic-embed-text:latest
ollama pull llama3:8b
```

### 5. Instalează PostgreSQL (opțional, dacă folosești DB)

```bash
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

## Configurare

### 6. Creează fișierul .env

```bash
cd rag
nano .env
```

Conținut `.env`:
```env
# Database (opțional)
DB_URL=postgres://user:password@localhost:5432/rag_db

# RAG Configuration
RAG_PROVIDER=ollama
RAG_INDEX_PATH=/path/to/rag/data/rag_index.json
RAG_OLLAMA_BASE_URL=http://localhost:11434
RAG_EMBEDDING_MODEL=nomic-embed-text:latest
RAG_CHAT_MODEL=llama3:8b
```

### 7. Creează directoarele necesare

```bash
mkdir -p data
mkdir -p docs
```

## Build și Run

### 8. Build aplicația

```bash
# Instalează dependențele
go mod download

# Build binarul
go build -o rag-server ./cmd/main.go

# Sau build pentru producție (optimizat)
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o rag-server ./cmd/main.go
```

### 9. Rulează ingestion-ul (prima dată)

```bash
# Rulează ingestion pentru a crea index-ul
go run ./cmd/rag --mode ingest --index data/rag_index.json
```

### 10. Testează aplicația

```bash
# Rulează serverul
./rag-server

# Sau
go run cmd/main.go
```

Aplicația ar trebui să ruleze pe `http://localhost:8000`

## Deployment Production

### 11. Creează Systemd Service

```bash
sudo nano /etc/systemd/system/rag.service
```

Conținut:
```ini
[Unit]
Description=RAG Application
After=network.target postgresql.service ollama.service

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/rag
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
ExecStart=/path/to/rag/rag-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Activează serviciul:
```bash
sudo systemctl daemon-reload
sudo systemctl enable rag
sudo systemctl start rag
sudo systemctl status rag
```

### 12. Configurează Nginx (Reverse Proxy)

```bash
sudo nano /etc/nginx/sites-available/rag
```

Conținut:
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

Activează:
```bash
sudo ln -s /etc/nginx/sites-available/rag /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 13. SSL cu Let's Encrypt (opțional)

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## Comenzi Utile

### Verificare status
```bash
# Status aplicație
sudo systemctl status rag

# Logs aplicație
sudo journalctl -u rag -f

# Status Ollama
sudo systemctl status ollama

# Verifică portul
netstat -tulpn | grep 8000
```

### Restart
```bash
sudo systemctl restart rag
sudo systemctl restart ollama
```

### Update aplicație
```bash
cd /path/to/rag
git pull origin main
go build -o rag-server ./cmd/main.go
sudo systemctl restart rag
```

## Troubleshooting

### Ollama nu răspunde
```bash
# Verifică dacă rulează
sudo systemctl status ollama

# Repornește
sudo systemctl restart ollama

# Verifică modelele
ollama list
```

### Port 8000 ocupat
```bash
# Găsește procesul
sudo lsof -i :8000

# Oprește procesul
sudo kill -9 <PID>
```

### Permisiuni
```bash
# Asigură-te că user-ul are acces la directoare
sudo chown -R your-user:your-user /path/to/rag
chmod +x rag-server
```

## Variabile de Mediu Importante

- `RAG_INDEX_PATH` - Calea către index-ul RAG (absolută)
- `RAG_PROVIDER` - `ollama` sau `openai`
- `RAG_OLLAMA_BASE_URL` - URL-ul Ollama (default: http://localhost:11434)
- `DB_URL` - Connection string PostgreSQL (opțional)

## Note

- Asigură-te că Ollama rulează înainte de a porni aplicația
- Index-ul RAG trebuie creat înainte de prima utilizare
- Pentru producție, folosește HTTPS
- Monitorizează resursele (CPU, RAM) - Ollama poate consuma multă memorie

