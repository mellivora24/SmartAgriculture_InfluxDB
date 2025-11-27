# React + TypeScript + Vite

This template provides a minimal setup to get React working in Vite with HMR and some ESLint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Babel](https://babeljs.io/) (or [oxc](https://oxc.rs) when used in [rolldown-vite](https://vite.dev/guide/rolldown)) for Fast Refresh
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/) for Fast Refresh

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the ESLint configuration

If you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```js
export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...

      // Remove tseslint.configs.recommended and replace with this
      tseslint.configs.recommendedTypeChecked,
      // Alternatively, use this for stricter rules
      tseslint.configs.strictTypeChecked,
      // Optionally, add this for stylistic rules
      tseslint.configs.stylisticTypeChecked,

      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```


# CAU TRUC DU AN
```bash
src/
├── main.tsx
├── App.tsx
├── core/                     # Nền tảng chung
│   ├── api/                  # Axios client, interceptors
│   ├── config/               # Env, constants
│   ├── hooks/                # Custom hook toàn cục
│   └── utils/                # Helper functions
├── features/                 # Feature riêng từng màn hình / chức năng
│   ├── auth/                 # Đăng nhập / đăng ký
│   │   ├── api/              # call API login/register
│   │   ├── components/       # LoginForm, RegisterForm
│   │   ├── hooks/            # useAuth, useToken
│   │   └── pages/
│   │       ├── LoginPage.tsx
│   │       └── RegisterPage.tsx
│   ├── onboarding/           # Màn hình onboarding
│   │   ├── components/
│   │   └── pages/
│   │       └── OnboardingPage.tsx
│   ├── dashboard/            # Trang tổng quan realtime
│   │   ├── api/              # Gọi farm, MCU, surveyPoint, sensorData
│   │   ├── components/       # CardFarm, CardMCU, SensorChart, CommandCard
│   │   ├── hooks/            # useDashboardData, useRealtimeUpdates
│   │   └── pages/
│   │       └── DashboardPage.tsx
│   ├── surveyPoint/          # Quản lý survey point
│   │   ├── api/
│   │   ├── components/       # SurveyPointForm, SurveyPointTable
│   │   ├── hooks/            # useSurveyPoint
│   │   └── pages/
│   │       └── SurveyPointManagementPage.tsx
│   ├── deviceCommand/        # Lịch sử bật/tắt thiết bị
│   │   ├── api/
│   │   ├── components/       # CommandTable
│   │   ├── hooks/            # useDeviceCommands
│   │   └── pages/
│   │       └── DeviceCommandHistoryPage.tsx
│   ├── sensorData/           # Lịch sử dữ liệu cảm biến
│   │   ├── api/
│   │   ├── components/       # SensorTable, SensorChart
│   │   ├── hooks/            # useSensorData
│   │   └── pages/
│   │       └── SensorDataHistoryPage.tsx
│   ├── account/              # Cài đặt tài khoản
│   │   ├── api/              # Update, delete, change password
│   │   ├── components/       # AccountForm, AvatarUploader
│   │   ├── hooks/            # useAccount
│   │   └── pages/
│   │       └── AccountSettingsPage.tsx
│   └── media/                # Lưu ảnh dùng bên thứ 3
│       ├── api/              # Upload image API / Cloudinary
│       └── hooks/            # useMedia
├── shared/                   # Component dùng chung
│   ├── components/           # Button, Modal, Input, Table...
│   └── ui/                   # Theme, Typography, Layout
├── assets/                   # Ảnh, svg, fonts
└── styles/                   # CSS global / Tailwind
```