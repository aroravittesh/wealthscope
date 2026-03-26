# 🚀 WealthScope – Stock Portfolio Intelligence Platform

## 👥 Team Members

* Vittesh Arora – Backend
* Ansh Jain – Backend
* Raghav Gupta – Frontend
* Rishithaa Maligireddy – Frontend

---

## 📌 Project Overview

WealthScope is a full-stack Stock Portfolio Intelligence Platform designed to demonstrate real-world software engineering practices including system design, modular architecture, database management, machine learning integration, and cloud deployment.

The platform allows users to manage portfolios, track holdings, and analyze performance through a scalable and maintainable architecture.

---

## 🧠 Problem Statement

Most portfolio tools lack transparency and many academic projects lack real-world architecture and deployment practices. WealthScope addresses this by building a clean, modular, and production-style system.

---

## 🎯 Objectives

* Build a full-stack system using modern technologies
* Develop REST APIs in Go with clean architecture
* Design a relational database
* Integrate ML services
* Implement CI/CD and deployment

---

## ⚙️ Tech Stack

* Frontend: Angular, TypeScript
* Backend: Go (Golang), Gorilla Mux
* Database: PostgreSQL
* ML: Python
* DevOps: Docker, CI/CD

---

## 🏗️ Architecture

Frontend (Angular) → Backend (Go APIs) → Database (PostgreSQL) → ML Service (Python)

---

## 🔑 Features

* User authentication (JWT)
* Portfolio management
* Holdings tracking
* Portfolio analytics
* ML-based insights

---

## 🚀 Setup Instructions

### Backend

```bash
git clone https://github.com/aroravittesh/wealthscope.git
cd wealthscope/backend
go mod tidy
go run main.go
```

### Frontend

```bash
cd Frontend
npm install
ng serve
```

### Database

* Setup PostgreSQL
* Run schema SQL

---

## 🧪 Testing

* Backend: unit tests using Go testing
* Frontend: tested using Cypress for login, portfolio, and holdings flows

Run backend tests:

```bash
go test ./...
```

Run frontend tests:

```bash
cd Frontend
npx cypress open
```

---

## 📌 API Base URL

http://localhost:8080/api

---

## 📦 Scope

Included:

* Portfolio & holdings management
* Analytics & ML integration

Excluded:

* Real-time trading
* Financial advice

---

## ⚠️ Disclaimer

This project is for academic purposes only and does not provide financial advice.

---

## 📌 Contributors

* https://github.com/aroravittesh
* https://github.com/leo-Ansh2004
* https://github.com/raghhavv03
* https://github.com/Rishithaa-88

---

## ⭐ Conclusion

WealthScope demonstrates a complete full-stack system with scalable architecture, integrating backend, frontend, and analytics into a single platform.
