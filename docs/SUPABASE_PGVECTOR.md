# Supabase pgvector - Vector Store în PostgreSQL

## Ce este pgvector?

pgvector este o extensie PostgreSQL care permite stocarea și căutarea vectorilor (embeddings) direct în baza de date.

## Avantaje

- ✅ **Scalabil** - până la milioane de vectori
- ✅ **Performant** - index HNSW pentru căutări rapide
- ✅ **Integrat** - folosește deja Supabase
- ✅ **Backup automat** - inclus în Supabase
- ✅ **Free tier** - 500MB database storage

## Setup

### 1. Activează pgvector în Supabase

În Supabase Dashboard:
1. Mergi la **Database** → **Extensions**
2. Caută **pgvector**
3. Click **Enable**

Sau rulează SQL:
```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### 2. Creează Tabelul pentru Vectori

```sql
CREATE TABLE rag_chunks (
    id BIGSERIAL PRIMARY KEY,
    chunk_id TEXT NOT NULL,
    document_id TEXT NOT NULL,
    title TEXT NOT NULL,
    uri TEXT NOT NULL,
    source TEXT NOT NULL,
    text TEXT NOT NULL,
    embedding vector(768), -- Dimensiunea embedding-ului (768 pentru nomic-embed-text)
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index HNSW pentru căutări rapide
CREATE INDEX ON rag_chunks USING hnsw (embedding vector_cosine_ops);

-- Index pentru document_id (pentru filtrare)
CREATE INDEX idx_rag_chunks_document_id ON rag_chunks(document_id);
```

### 3. Dimensiunea Embedding-ului

Verifică dimensiunea embedding-ului modelului tău:
- **nomic-embed-text**: 768 dimensiuni
- **text-embedding-ada-002** (OpenAI): 1536 dimensiuni
- **text-embedding-3-small** (OpenAI): 1536 dimensiuni

Ajustează `vector(768)` în funcție de modelul tău.

### 4. Implementare în Go

Trebuie să creezi un nou `VectorStore` care folosește PostgreSQL în loc de JSON.

**Structuri:**
```go
type PGVectorStore struct {
    db *gorm.DB
    embeddingDim int
}

type RAGChunk struct {
    ID         uint      `gorm:"primaryKey"`
    ChunkID    string    `gorm:"index"`
    DocumentID string    `gorm:"index"`
    Title      string
    URI        string
    Source     string
    Text       string
    Embedding  []float32 `gorm:"type:vector"`
    Metadata   datatypes.JSON
    CreatedAt  time.Time
}
```

**Căutare cu cosine similarity:**
```sql
SELECT 
    chunk_id, document_id, title, uri, source, text,
    1 - (embedding <=> $1::vector) as similarity
FROM rag_chunks
ORDER BY embedding <=> $1::vector
LIMIT $2;
```

## Migrare de la JSON la PostgreSQL

1. **Backup index-ul actual:**
   ```bash
   cp data/rag_index.json data/rag_index.json.backup
   ```

2. **Script de migrare:**
   - Citește `rag_index.json`
   - Inserează toate chunk-urile în PostgreSQL
   - Verifică că toate au fost inserate

3. **Testează:**
   - Rulează câteva query-uri
   - Compară rezultatele cu JSON-ul vechi

## Performanță

**Comparație:**
- **JSON file (10k chunks)**: ~500ms pentru search
- **pgvector (10k chunks)**: ~50ms pentru search
- **pgvector (100k chunks)**: ~100ms pentru search
- **pgvector (1M chunks)**: ~200ms pentru search

**Optimizări:**
- Folosește index HNSW
- Ajustează parametrii HNSW pentru trade-off memorie/viteză
- Folosește connection pooling

## Costuri

**Supabase Free Tier:**
- 500MB database storage
- ~50k chunks (cu embedding 768 dim) = ~200MB
- Sufficient pentru început

**Supabase Pro ($25/month):**
- 8GB database storage
- ~2M chunks = ~8GB
- Sufficient pentru majoritatea cazurilor

## Backup

Supabase face backup automat, dar poți exporta manual:
```sql
-- Export to CSV
COPY rag_chunks TO '/tmp/rag_chunks_backup.csv' CSV HEADER;
```

## Rollback

Dacă vrei să revii la JSON:
1. Exportă din PostgreSQL
2. Reconstruiește `rag_index.json`
3. Modifică codul să folosească JSON din nou

