# 🚀 Rishuu Frontend - Professional Portfolio Management Dashboard

A modern, responsive Angular 19+ fintech frontend application for smart portfolio management with beautiful UI/UX design.

## ✨ Features

- **Authentication**: User login, registration, and JWT token management
- **Dashboard**: Real-time portfolio metrics and analytics
- **Portfolio Management**: Create, view, and manage investment portfolios
- **Asset Tracking**: Monitor holdings with real-time profit/loss calculations
- **Beautiful UI**: Dark theme with glassmorphism effects using Tailwind CSS
- **Responsive Design**: Mobile-first, works on all devices
- **Type-Safe**: Full TypeScript support with interfaces
- **Modern Stack**: Angular 19, RxJS, Tailwind CSS

## 📋 Project Structure

```
src/
├── app/
│   ├── core/                 # Core services & interceptors
│   ├── features/            # Feature modules
│   │   ├── auth/           # Authentication (login, register)
│   │   ├── dashboard/      # Dashboard components
│   │   ├── portfolio/      # Portfolio management
│   │   ├── transactions/   # Transaction history
│   │   └── admin/          # Admin panel (future)
│   ├── shared/             # Shared components (navbar, etc)
│   ├── services/           # API services
│   ├── guards/             # Route guards
│   ├── interceptors/       # HTTP interceptors
│   ├── models/             # TypeScript interfaces
│   └── layouts/            # Layout components
├── assets/                 # Static assets
└── styles.scss             # Global styles
```

## 🛠️ Tech Stack

- **Frontend Framework**: Angular 19+
- **Language**: TypeScript 5+
- **Styling**: Tailwind CSS 3.4 + SCSS
- **State Management**: RxJS Observables
- **HTTP Client**: Axios + Angular HttpClient
- **Charts**: Chart.js (ng2-charts)
- **Build Tool**: Angular CLI + esbuild

## 🚀 Getting Started

### Prerequisites

- Node.js 18+ (v23.11.0 tested)
- npm 10+

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/Rishuu-Frontend.git
cd Rishuu-Frontend
```

2. **Install dependencies**
```bash
npm install --legacy-peer-deps
```

3. **Start development server**
```bash
npm start
```

The application will be available at `http://localhost:4200`

### Build for Production

```bash
npm run build
```

Output will be in `dist/rishuu-frontend/`

## 📦 Available Scripts

- `npm start` - Start development server (ng serve)
- `npm run build` - Create production build
- `npm run watch` - Build with watch mode
- `npm test` - Run unit tests (ng test)

## 🔐 Authentication

The app uses JWT (JSON Web Tokens) for authentication:

- **Login/Register**: POST to `/api/auth/login` and `/api/auth/register`
- **Token Storage**: Tokens stored in localStorage
- **Auto Logout**: Automatic logout on token expiration
- **HTTP Interceptor**: Automatically adds JWT to requests

## 🎨 Design System

### Color Palette
- **Primary**: Blue (#3B82F6)
- **Secondary**: Purple (#8B5CF6)
- **Dark**: Slate-900 (#0F172A)
- **Success**: Green (#10B981)
- **Warning**: Amber (#F59E0B)
- **Danger**: Red (#EF4444)

### Components
- Glass effect cards with backdrop blur
- Smooth animations and transitions
- Responsive grid system
- Icons from inline SVG

## 📱 Responsive Breakpoints

- **Mobile**: < 768px
- **Tablet**: 768px - 1024px
- **Desktop**: > 1024px

## 🔌 API Integration

Backend API endpoints (configure in services):

```typescript
// Base URL: http://localhost:8080/api

// Auth endpoints
POST   /auth/register
POST   /auth/login
POST   /auth/refresh
POST   /auth/logout

// Portfolio endpoints
GET    /portfolios
GET    /portfolios/:id
POST   /portfolios
PUT    /portfolios/:id
DELETE /portfolios/:id
GET    /portfolios/:id/metrics
GET    /portfolios/:id/holdings
```

## 🔄 State Management

Uses RxJS BehaviorSubjects for state management:

```typescript
// AuthService
authService.isAuthenticated$ // Observable<boolean>
authService.currentUser$     // Observable<User | null>

// PortfolioService
portfolioService.portfolios$ // Observable<Portfolio[]>
portfolioService.metrics$    // Observable<DashboardMetrics | null>
```

## 📚 Component Architecture

### Smart Components (Containers)
- `DashboardOverviewComponent`
- `LoginComponent`
- `RegisterComponent`

### Presentation Components
- `NavbarComponent`
- Portfolio cards
- Metric cards

## 🧪 Testing

```bash
# Run unit tests
npm test

# Run tests with coverage
ng test --code-coverage
```

## 🚀 Deployment

### Build
```bash
npm run build
```

### Deploy to Netlify/Vercel
```bash
# Netlify
npm install -g netlify-cli
netlify deploy --prod --dir=dist/rishuu-frontend

# Vercel
npm install -g vercel
vercel --prod
```

### Docker (Optional)
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
RUN npm install --legacy-peer-deps
RUN npm run build
EXPOSE 4200
CMD ["npm", "start"]
```

## 📝 Environment Configuration

Create `.env` file:

```
ANGULAR_APP_API_URL=http://localhost:8080/api
ANGULAR_APP_ENV=development
```

Access in code:

```typescript
import { environment } from './environments/environment';
environment.apiUrl
```

## 🔒 Security Best Practices

- ✅ JWT token-based authentication
- ✅ HttpOnly cookies (configurable)
- ✅ CSRF protection ready
- ✅ XSS protection via Angular sanitizer
- ✅ Secure password validation
- ✅ Route guards for protected pages
- ✅ Automatic token refresh
- ✅ Logout on 401 Unauthorized

## 📖 File Descriptions

### Services
- **auth.service.ts**: User authentication & token management
- **portfolio.service.ts**: Portfolio data operations

### Interceptors
- **auth.interceptor.ts**: Adds JWT to requests, handles 401 errors

### Guards
- **auth.guard.ts**: Protects authenticated routes

### Models
- **index.ts**: All TypeScript interfaces

## 🎯 Development Workflow

1. Create a new branch: `git checkout -b feature/feature-name`
2. Make your changes
3. Commit: `git commit -am 'Add feature'`
4. Push: `git push origin feature/feature-name`
5. Open a Pull Request

## 📊 Performance

- Angular Standalone components for tree-shaking
- OnPush change detection strategy
- RxJS unsubscribe patterns with `takeUntil`
- Lazy loading routes (ready to implement)
- CSS-in-JS optimization via Tailwind

## 🐛 Troubleshooting

### Port 4200 already in use
```bash
npm start -- --port 4300
```

### Module not found errors
```bash
rm -rf node_modules
npm install --legacy-peer-deps
```

### Build errors
```bash
ng clean
npm run build
```

## 📞 Support

For issues and questions:
- Check existing GitHub issues
- Create a new issue with details
- Include error logs and steps to reproduce

## 📄 License

MIT License - feel free to use this project

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

---

**Happy Coding! 🎉**

Made with ❤️ for portfolio management
