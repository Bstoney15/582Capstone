import { createBrowserRouter, RouterProvider, Outlet, Navigate } from "react-router-dom";
import { ThemeProvider } from "./contexts/ThemeContext";
import { AuthProvider } from "./contexts/AuthContext";
import ProtectedRoute from "./components/ProtectedRoute";
import Dashboard from "./pages/Dashboard/Dashboard.jsx";
import Profile from "./pages/Profile/Profile.jsx";
import Settings from "./pages/Settings/Settings.jsx";
import Home from "./pages/Home/Home.jsx";
import Login from "./pages/Login/Login.jsx";
import Signup from "./pages/Signup/Signup.jsx";
import Webhooks from "./pages/Webhooks/Webhooks.jsx";
import ApiKeys from "./pages/ApiKeys/ApiKeys.jsx";
import Wallet from "./pages/Wallet/Wallet.jsx";
import Users from "./pages/Users/Users.jsx";
import "./App.css";

const router = createBrowserRouter([
  {
    element: <Outlet />,
    children: [
      {
        path: "/",
        element: <Home />
      },
      {
        path: "/login",
        element: <Login />
      },
      {
        path: "/signup",
        element: <Signup />
      },
      {
        element: <ProtectedRoute />,
        children: [
          {
            path: "/dashboard",
            element: <Dashboard />
          },
          {
            path: "/profile",
            element: <Profile />
          },
          {
            path: "/settings",
            element: <Settings />
          },
          {
            path: "/webhooks",
            element: <Webhooks />
          },
          {
            path: "/api-keys",
            element: <ApiKeys />
          },
          {
            path: "/wallet",
            element: <Wallet />
          },
          {
            path: "/users",
            element: <Users />
          }
        ]
      },
      {
        path: "*",
        element: <Navigate to="/" replace />
      }
    ]
  }
]);

export default function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </ThemeProvider>
  );
}