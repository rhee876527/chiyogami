## Chiyogami

#### Chiyogami is a sleek, modern pastebin with encryption, customizable expiry, private pastes, user accounts and an API for developers.




##### Screenshots:

![screen-1](https://github.com/user-attachments/assets/5985f94d-4e35-4479-bc57-726e7cfb4577)
![screen-2](https://github.com/user-attachments/assets/0918a641-bf50-4d26-971a-39d7e9876a6d)
![screen-3](https://github.com/user-attachments/assets/95532b56-9e2f-447f-8c9c-cdbe4119fa59)





✨ **Features**

- ✔ Beautiful & Responsive UI — Built with TailwindCSS & DaisyUI for a clean and modern look.
- 🖍 Syntax Highlighting — Automatic formatting with HighlightJS.
- 📝 Markdown Rendering — Automatic formatting with Marked.
- ⏳ Configurable Expiry — Set custom expiration times with API.
- 🔒 Secure & Private — Client-side encryption with WebCryptoAPI for encrypted pastes. No password saved in server.
- 📡 Powerful API — Create and fetch pastes without leaving the terminal.
- 🔍 Public Pastes — List & search all public pastes.
- 🔑 Private Pastes — Only accessible via a unique, unguessable link for enhanced privacy (use encryption on web UI for ultimate privacy).
- 🗄 Local Storage — Uses SQLite for a lightweight, self-hostable database.
- 👤 User Accounts — Create & manage your pastes with authentication.
- 🔗 Easy Sharing — Share paste links or scan a QR code for instant access.
- 🛡 Built-in Rate Limiting — Protects against spam and abuse with smart request throttling.
- 🐳 Easy self-host with docker.




## Installation
Docker. Build it or check [docker-compose](https://github.com/rhee876527/chiyogami/blob/main/docker-compose.yml) file for example with pre-built images.

### Quick run

```
 docker run -d \
  -v "$(pwd)/pastes:/pastes" \
  -p 127.0.0.1:8000:8000 \
  --restart unless-stopped \
  ghcr.io/rhee876527/chiyogami:latest
```

### Environment variables


| Env     | Default value | Note   |
|----------|-----|--------------|
| SECRET_KEY    | None  |  Session key. Keep secure and unique. Recommended example: openssl rand -base64 32.    |
| PASTE_DEFAULT_EXPIRATION      | 24h  |    Valid time units are “ns”, “us” (or “µs”), “ms”, “s”, “m”, “h” or `"Never"`  |
| MAX_CHAR_CONTENT  | 50000  | Maximum characters allowed in content body. Larger limits may offer reduced performance     |
| DISABLE_RATE_LIMIT  | None  | Use `1` to disable rate limit     |
| DATABASE_PATH  | None  | For local development use. Conflicts with docker volume paths.      |


## Usage
Web UI is simple & straightforward. Or use the `API`.

#### Create paste
***FILE***
```
curl http://localhost:8000/paste -T sample.txt
```
or **JSON**
```
curl -X POST \
  http://localhost:8000/paste \
  -H 'Content-Type: application/json' \
  -d '{"content":"Test paste"}'
```




**response:** `{"title":"OkxI"}`




Note: Pastes are created by default with `Public` `visibility`. They can be accessed from api or website.
Change this to `Private` or `Unlisted` to make the paste undiscoverable. Pastes are also set to expire within 24hrs if expiry is not specified.

You can set a default expiry for new pastes with `PASTE_DEFAULT_EXPIRATION`.




#### Create private paste with 48h expiry

```
curl -X POST \
  http://localhost:8000/paste \
  -H 'Content-Type: application/json' \
  -d '{"content":"Test", "visibility":"Private", "expiration":"48h"}'
```
**response:** `{"title":"euVa"}`




#### Fetch created paste
```
curl -X GET http://localhost:8000/paste/bZTR -H "Accept: application/json"
```



**response:**
``
{"ID":22,"CreatedAt":"2025-02-04T19:48:06.747679947Z","UpdatedAt":"2025-02-04T19:48:06.747679947Z","DeletedAt":null,"Title":"bZTR","Content":"test private","Visibility":"Private","expiration":"2025-02-05T19:48:06.747635027Z","IsEncrypted":false,"UserID":0,"IsUserPaste":false}
``



#### Create paste from file
```
f=insert*file*name; \
    jq -Rs '{content: .}' < "$f" | \
    curl -X POST http://localhost:8000/paste \
     -H 'Content-Type: application/json' \
     -d @-
```



**response:** `{"title":"awDI"}`




#### Delete owner paste using session (from cookies)

```
curl -X DELETE http://localhost:8000/paste/EIKq \
-b "session=MTczNzA2NDI5NXxEWDhFQVFMX2dBQUJFQUVRQUFBZl80QUFBUVp6ZEhKcGJtY01DUUFIZFhObGNsOXBaQVIxYVc1MEJnSUFEQT09fLnhi2OxsN6coY5ZmmBeA0tPXUcsKiii6ECOoJ7yrqNC"
```

**response:** `{"message":"Paste deleted successfully"}`




##### COPYRIGHT
This software is free to use in accordance with the [license](https://github.com/rhee876527/chiyogami/blob/main/LICENSE).
