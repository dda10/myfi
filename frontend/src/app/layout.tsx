import type { Metadata } from "next";
import { Geist, Geist_Mono, Inter } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/context/ThemeContext";
import { I18nProvider } from "@/context/I18nContext";
import { AuthProvider, ProtectedRoute } from "@/context/AuthContext";
import { AuthenticatedLayout } from "@/components/layout/AuthenticatedLayout";
import { ErrorBoundary } from "@/components/common/ErrorBoundary";
import { ServiceWorkerRegistrar } from "@/components/common/ServiceWorkerRegistrar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700", "800", "900"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "MyFi — AI-Powered Investment Platform",
  description: "Vietnamese Stock Analysis & AI-Powered Investment Platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        suppressHydrationWarning
        className={`${inter.variable} ${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <AuthProvider>
          <ThemeProvider>
            <I18nProvider>
              <ProtectedRoute>
                <ErrorBoundary>
                <ServiceWorkerRegistrar />
                <AuthenticatedLayout>
                  {children}
                </AuthenticatedLayout>
                </ErrorBoundary>
              </ProtectedRoute>
            </I18nProvider>
          </ThemeProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
