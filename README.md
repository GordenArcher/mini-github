### `README.md`

````markdown
# Mini-GitHub

[![Go](https://img.shields.io/badge/Go-1.21-blue?logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-blue?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-orange?logo=redis)](https://redis.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A lightweight GitHub-like platform built with **Go**, **Gin**, **PostgreSQL**, **Redis**, and **Git**.  
Users can register, create repositories, push code, and manage projects locally. Ideal for learning, experimentation, and personal projects.

---

## Features

- User authentication with JWT (register, login, email verification, password reset)
- Repository management (create, list, get details)
- Bare Git repository creation for each user
- Push and pull code to repositories using standard Git commands
- Email notifications for verification and password reset
- Rate limiting middleware

---

## Tech Stack

- **Backend:** Go, Gin, GORM  
- **Database:** PostgreSQL  
- **Cache / Queue:** Redis  
- **Mailer:** SMTP (Gmail)  
- **Version Control:** Git  

---

## Setup

1. **Clone the repository**

```bash
git clone github.com:GordenArcher/mini-github.git
cd mini-github
````

2. **Create a `.env` file** with your configuration:

```env
SERVER_PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/dbname
REDIS_ADDR=localhost:6379
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=youremail@gmail.com
SMTP_PASS=yourapppassword
JWT_ACCESS_SECRET=youraccesstokensecret
JWT_REFRESH_SECRET=yourrefreshtokensecret
```

3. **Run database migrations**

```bash
go run cmd/server/main.go
```

4. **Start the server**

```bash
go run cmd/server/main.go
```

---

## API Endpoints

### Auth

| Method | Endpoint                             | Description                  |
| ------ | ----------------------------------   | ---------------------------- |
| POST   | `/api/v1/auth/register`              | Register a new user          |
| GET    | `/api/v1/auth/verify`                | Verify email with token      |
| POST   | `/api/v1/auth/resend-verification`   | Resend verification email    |
| POST   | `/api/v1/auth/login`                 | Login and receive JWT tokens |
| POST   | `/api/v1/auth/refresh`               | Refresh access token         |
| POST   | `/api/v1/auth/logout`                | Logout and revoke tokens     |
| POST   | `/api/v1/auth/request-password-reset`| send mail with a token       |
| POST   | `/api/v1/auth/reset-password`        | Resets user password         |

### Repositories

| Method | Endpoint               | Description                    |
| ------ | ---------------------- | ------------------------------ |
| POST   | `/api/v1/repos/create` | Create a new repository (bare) |
| GET    | `/api/v1/repos/`       | List all user repositories     |
| GET    | `/api/v1/repos/:id`    | Get repository details         |

---

## Using Git with Your Repositories

After creating a repository via the API, you can push your local code to it:

1. **Add the remote**

```bash
git remote add origin /path/to/mini-github-repos/<user_id>/<repo_name>.git
```

2. **Make your first commit**

```bash
git add .
git commit -m "Initial commit"
```

3. **Push to your remote repository**

```bash
git push -u origin main
```

> Note: Make sure your branch matches the one you’re pushing (`main` or `master`).

---

## Folder Structure

```
cmd/server       # Entry point
internal/config      # Configurations for the project
internal/db      # Database models and connection
internal/errors      # Error handling
internal/handlers # Gin handlers for auth and repos
internal/middleware # JWT auth, rate limiting
internal/routes   # API route definitions
internal/mail     # Mailer utility
internal/redis    # Redis client helpers
internal/log      # Logger setup
tests      # Test unit for auth
.env.example # How the environment variables structure looks like 
```

---

## License

This project is licensed under the MIT License.

```
