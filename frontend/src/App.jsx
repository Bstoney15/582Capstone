// App.jsx – root application component that defines the client-side router and wraps the app in global context providers.
// Author: Ben Stonestreet, Connor Williamson, Charley Finling, Ryan Grimsley
// Created: 2026-02-02

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
import WidgetDemo from "./pages/WidgetDemo/WidgetDemo.jsx";
import Webhooks from "./pages/Webhooks/Webhooks.jsx";
import ApiKeys from "./pages/ApiKeys/ApiKeys.jsx";
import CreateMerchant from "./pages/CreateMerchant/CreateMerchant.jsx";
import Wallet from "./pages/Wallet/Wallet.jsx";
import Users from "./pages/Users/Users.jsx";
import Customers from "./pages/Customers/Customers.jsx";
import ApiDocs from "./pages/ApiDocs/ApiDocs.jsx";
import "./App.css";

// router defines all application routes. Protected routes are nested under ProtectedRoute,
// which enforces authentication and renders the shared NavBar.
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
        path: "/widget-demo",
        element: <WidgetDemo />
      },

      {
        element: <ProtectedRoute />,
        children: [
          {
            path: "/create-merchant",
            element: <CreateMerchant />
          },
          {
            path: "/api-docs",
            element: <ApiDocs />
          },
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
          },
          {
            path: "/customers",
            element: <Customers />
          },
          {
            path: "/create-merchant",
            element: <CreateMerchant />
          },
        ]
      },
      {
        path: "*",
        element: <Navigate to="/" replace />
      }
    ]
  }
]);

/**
 * App is the root component. It wraps the entire application in ThemeProvider and
 * AuthProvider so that theme and auth state are available to every descendant.
 * @returns {JSX.Element} The rendered application with router.
 */
export default function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </ThemeProvider>
  );
}
