// Package handler contains HTTP request handlers for the API endpoints.
// Handlers are responsible for parsing requests, calling services, and
// formatting responses.
//
// Endpoints:
//   - POST /api/views/:slug     - Record a page view
//   - GET  /api/views/:slug     - Get view count for a post
//   - GET  /api/views           - Get bulk view counts
//   - POST /api/newsletter/subscribe   - Subscribe to newsletter
//   - POST /api/newsletter/unsubscribe - Unsubscribe from newsletter
//   - GET  /api/newsletter/verify/:token - Verify email
//   - GET  /api/health          - Health check
package handler
