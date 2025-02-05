## Chiyogami 

Yet Another Pastebin

#### Chiyogami is a sleek, modern pastebin with encryption, customizable expiry, private pastes, user accounts and an API for developers. 🚀

<br>

✨ **Features**

- ✔ Beautiful & Responsive UI — Built with TailwindCSS & DaisyUI for a clean and modern look.
- 🖍 Syntax Highlighting — Automatic formatting with HighlightJS.
- ⏳ Configurable Expiry — Set custom expiration times with API.
- 🔒 Secure & Private — Client-side encryption with WebCrypto for encrypted pastes. No password saved in server.
- 📡 Powerful API — Create and fetch pastes without leaving the terminal.
- 🔍 Public Pastes — List & search all public pastes.
- 🔑 Private Pastes — Only accessible via a unique, unguessable link for enhanced privacy (use encryption on web UI for ultimate privacy).
- 🗄 Local Storage — Uses SQLite for a lightweight, self-hostable database.
- 👤 User Accounts — Create & manage your pastes with authentication.
- 🔗 Easy Sharing — Share paste links or scan a QR code for instant access.
- 🛡 Built-in Rate Limiting — Protects against spam and abuse with smart request throttling.
- 🚀 Easy self-host with powerful features!

<br><br>

## Usage
Web UI is simple & straightforward. Or use the `API`.

#### Create paste
``
curl -X POST \
  http://localhost:8000/paste \
  -H 'Content-Type: application/json' \
  -d '{"content":"Test paste"}'
{"title":"OkxI"}
``

<br>

Note: Pastes are created by default with `Public` `visibility`. They can be accessed from api or website.
Change this to `Private` or `Unlisted` to make the paste undiscoverable. Pastes also set to expire within 24hrs if expiry is not specified. You can set a default expiry for pastes created with `PASTE_DEFAULT_EXPIRATION`.

#### Fetch created paste
``
curl -X GET http://localhost:8000/paste/bZTR -H "Accept: application/json"
{"ID":22,"CreatedAt":"2025-02-04T19:48:06.747679947Z","UpdatedAt":"2025-02-04T19:48:06.747679947Z","DeletedAt":null,"Title":"bZTR","Content":"test private","Visibility":"Private","expiration":"2025-02-05T19:48:06.747635027Z","IsEncrypted":false,"UserID":0,"IsUserPaste":false}
``

#### Create paste from txt file content
``
f=insert*file*name*.ext; jq -Rs '{content: .}' < "$f" | curl -X POST http://localhost:8000/paste -H 'Content-Type: application/json' -d @-
{"title":"awDI"}
``

The API has even more features. But I'll let you find that out instead of filling up this README page!

<br><br>

## Installation
Docker. Build it or check [docker-compose](https://github.com/rhee876527/chiyogami/blob/main/docker-compose.yml) file example for pre-built images.

<br><br>

### Screenshots:

![screen-1](https://github.com/user-attachments/assets/5985f94d-4e35-4479-bc57-726e7cfb4577)
![screen-2](https://github.com/user-attachments/assets/0918a641-bf50-4d26-971a-39d7e9876a6d)
![screen-3](https://github.com/user-attachments/assets/95532b56-9e2f-447f-8c9c-cdbe4119fa59)

<br>

##### COPYRIGHT
This software is free to use in accordance with the [license](https://github.com/rhee876527/chiyogami/blob/main/LICENSE).
