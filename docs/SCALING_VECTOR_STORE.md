# Scalarea Vector Store-ului

## Problema

Când index-ul RAG (`data/rag_index.json`) devine prea mare:
- **Probleme de memorie** - încărcarea întregului index în RAM
- **Probleme de spațiu** - fișier JSON foarte mare
- **Performanță lentă** - căutări lente prin toate chunk-urile

## Soluții

### 1. ✅ Supabase cu pgvector (Recomandat)

**Avantaje:**
- PostgreSQL cu extensia pgvector
- Scalabil (până la milioane de vectori)
- Căutări rapide cu index HNSW
- Backup automat
- Free tier generos

**Cum funcționează:**
- Vectorii sunt stocați în PostgreSQL
- Căutările folosesc index-ul HNSW pentru performanță
- Nu mai încărci totul în memorie

### 2. Vector Databases Dedicat

**Opțiuni:**
- **Qdrant** - Open source, self-hosted sau cloud
- **Pinecone** - Managed service (plătit)
- **Weaviate** - Open source, GraphQL API
- **Milvus** - Open source, foarte performant

**Avantaje:**
- Optimizat pentru vector search
- Scalare foarte bună
- API-uri dedicate

**Dezavantaje:**
- Necesită serviciu separat
- Configurare mai complexă

### 3. Optimizări pentru Fișier JSON

**Strategii:**
- **Compresie** - salvează ca `.json.gz`
- **Chunking mai eficient** - reduce overlap-ul
- **Filtrare** - șterge chunk-uri vechi/irelevante
- **Split pe fișiere** - împarte index-ul în mai multe fișiere

**Avantaje:**
- Simplu de implementat
- Nu necesită servicii externe

**Dezavantaje:**
- Tot trebuie să încarci în memorie
- Limitări la dimensiuni mari

### 4. Cloud Storage + Lazy Loading

**Strategie:**
- Salvează index-ul în S3/Cloud Storage
- Încarcă doar chunk-urile necesare
- Cache local pentru chunk-uri frecvente

**Avantaje:**
- Scalabil
- Nu ocupă spațiu local

**Dezavantaje:**
- Latency pentru căutări
- Necesită implementare complexă

## Recomandare

**Pentru început:** Supabase cu pgvector
- Deja ai Supabase configurat
- Nu necesită servicii noi
- Scalabil și performant
- Free tier suficient pentru început

**Pentru producție la scară:** Qdrant sau Milvus
- Optimizat pentru vector search
- Performanță superioară
- Self-hosted sau cloud

## Dimensiuni Orientative

- **< 10 MB** - JSON file OK
- **10-100 MB** - JSON file OK, dar consideră optimizări
- **100 MB - 1 GB** - Recomandat: Supabase pgvector
- **> 1 GB** - Necesar: Vector database dedicat

## Implementare pgvector în Supabase

Vezi `docs/SUPABASE_PGVECTOR.md` pentru ghid complet.

