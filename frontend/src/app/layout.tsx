import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/context/ThemeContext";
import { I18nProvider } from "@/context/I18nContext";
import { CurrencyProvider } from "@/context/CurrencyContext";
import { AuthProvider, ProtectedRoute } from "@/context/AuthContext";
import { AuthenticatedLayout } from "@/components/layout/AuthenticatedLayout";
import { ErrorBoundary } from "@/components/common/ErrorBoundary";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "MyFi Dashboard",
  description: "Personal Finance & AI Advisor",
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
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <AuthProvider>
          <ThemeProvider>
            <I18nProvider>
              <CurrencyProvider>
              <ProtectedRoute>
                <ErrorBoundary>
                <AuthenticatedLayout>
                  {children}
                </AuthenticatedLayout>
                </ErrorBoundary>
              </ProtectedRoute>
              </CurrencyProvider>
            </I18nProvider>
          </ThemeProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
