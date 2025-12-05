**1. User Authentication and Profile:**

**Login/Signup Page:** Implement a secure login/signup page for users.

**User Profiles:** Allow users to create and customize profiles.

**2. Home Page:**

**Featured Locations:** Display popular or trending travel locations.

**Search Bar:** Enable users to search for specific places or regions.

**3. Image Sharing:**

**Upload Images:** Users should be able to upload and share images from their travel locations.

**Image Gallery:** Organize images in a visually appealing gallery for each location.

**Comments/Descriptions:** Allow users to add comments or descriptions to their images.

**4. Transaction System:**

**Payment Integration:** Implement a secure payment gateway for transactions.

**Transaction History:** Provide users with a history of their transactions.

**5. Live Streaming (Future Feature):**

**Live Stream Page:** When live streaming is introduced, create a dedicated page for it.

**Notifications:** Allow users to receive notifications when a live stream is available in their desired location.

**6. Ratings and Reviews:**

**Rating System:** Implement a rating system for both images and live streams.

**Reviews and Feedback:** Allow users to leave reviews and feedback for the content they purchase.

**7. Map Integration:**

**Geotagging:** Integrate geotagging to display the location of images on a map.
Interactive Map: Users can explore different locations through an interactive map.

**8. Revenue Generation:**

**Transaction Fees:** Charge a percentage as a transaction fee for each successful transaction.

**Subscription Model (Optional):** Consider offering subscription plans for additional features or perks.

**9. User Support:**

**Help Center:** Include a help center or FAQ section to address common user queries.

**Customer Support:** Provide a way for users to contact customer support for assistance.

**10. Responsiveness and Accessibility:**

**Mobile Optimization:** Ensure the app is responsive and accessible on various devices.

**User-Friendly Design:** Prioritize a clean and intuitive design for a positive user experience.

**11. Marketing and Promotion:**

**Social Media Integration:** Allow users to share their travel experiences on social media.

**Referral Program (Optional):** Consider implementing a referral program to attract more users.

**12. Legal Considerations:**

**Terms of Service and Privacy Policy:** Clearly outline terms of service and privacy policies to protect both users and the platform.


Setup: https://www.youtube.com/watch?v=pbcTa-a3LBw

---

## Amazon SP-API Retrieval-Augmented Generation (RAG)

Engineers can now query the project knowledge base—local docs plus the official Amazon references shared in the brief—through a lightweight Retrieval-Augmented Generation (RAG) workflow.

### Prerequisites
- Create a `.env` file (or export in your shell) with:
  - `OPENAI_API_KEY=<your key>`
  - Optional overrides: `RAG_INDEX_PATH`, `RAG_CHAT_MODEL`, `RAG_EMBEDDING_MODEL`, `RAG_DEFAULT_TOP_K`.
- Ensure the `docs/` folder contains any internal notes you want embedded. Remote sources already include:
  - Amazon Selling Partner API samples README
  - Official SP-API rate limit guide + docs portal
  - Pilot/feature-toggle Google Sheet (TSV export)
  - plentymarkets `mc-amazon` repositories listing

### Build the vector store
Run the ingestion CLI, which fetches + chunks all sources, generates embeddings, and writes `data/rag_index.json`:
```
go run ./cmd/rag --mode ingest --index data/rag_index.json
```
You can point `--docs` to an alternate folder or tweak chunk sizing via `--chunk-size` / `--chunk-overlap`.

### Ask questions locally
```
go run ./cmd/rag --mode query --index data/rag_index.json \
  --question "How should we throttle SP-API calls for FBA orders?"
```
The CLI prints the synthesized answer plus the supporting sources/scores.

### API endpoint
Once an index exists, the Fiber server automatically wires `/api/rag/query`:
```
POST /api/rag/query
{
  "question": "What are the SP-API rate limit tiers?",
  "topK": 4     // optional override
}
```
If the service cannot load (missing key or index), the endpoint returns `503` with guidance.

### Regenerating data
The generated embeddings live under `data/` (git-ignored). Re-run the ingestion command whenever you add docs or when Amazon updates their public guidance.