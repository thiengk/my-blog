# 🎙️ Speech to Text Tool

Upload file MP3 (< 60 giây) và nhận transcript text ngay trên trình duyệt.

## Công nghệ

- **Astro 5** + **Svelte 5** + **TailwindCSS**
- **Transformers.js** (Whisper tiny model) — chạy hoàn toàn client-side
- Deploy: **Cloudflare Pages**

## Cách chạy

```bash
cd frontend
npm install
npm run dev
```

## Cách hoạt động

1. User upload file MP3/WAV/OGG (< 60s, < 10MB)
2. Browser tải model Whisper tiny (~40MB, cache sau lần đầu)
3. Audio được decode thành waveform 16kHz mono
4. Model Whisper xử lý và trả về transcript text
5. Không có dữ liệu nào rời khỏi trình duyệt

## Deploy

Push code lên branch `main` → Cloudflare Pages tự động build và deploy.
