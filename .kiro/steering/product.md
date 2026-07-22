# Product Overview

Tool chuyển đổi audio (MP3) thành text, chạy hoàn toàn trên trình duyệt sử dụng AI model Whisper.

## Tính năng chính

- **Upload Audio**: Kéo thả hoặc chọn file MP3/WAV/OGG (< 60s, < 10MB)
- **Speech-to-Text**: Sử dụng Whisper tiny model qua Transformers.js, xử lý hoàn toàn client-side
- **Privacy**: Dữ liệu không rời khỏi trình duyệt, không gửi lên server
- **Dark/Light Mode**: Detect system preference, toggle lưu localStorage
- **Copy kết quả**: Copy transcript text ra clipboard

## Kiến trúc

```
Cloudflare Pages → Static Frontend (Astro + Svelte + TailwindCSS)
                 → Whisper ONNX model (tải từ Hugging Face CDN, cached trong browser)
```

## Deployment

- Frontend: Cloudflare Pages (static, auto-build từ Git main branch)
