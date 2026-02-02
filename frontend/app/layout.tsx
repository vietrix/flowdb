import "./globals.css";
import type { ReactNode } from "react";
import Script from "next/script";
import { Providers } from "./providers";

export const metadata = {
  title: "DBAdmin",
  description: "Database administration workspace UI",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <head>
        <Script src="/flowdb-config.js" strategy="beforeInteractive" />
      </head>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
