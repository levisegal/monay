import type { Metadata } from "next";
import { Lora } from "next/font/google";
import "./globals.css";
import "../paper-portfolio.css";
import { Providers } from "./providers";

const lora = Lora({
  subsets: ["latin"],
  weight: ["400", "600"],
  variable: "--font-serif",
});

export const metadata: Metadata = {
  title: "Monay - Portfolio Dashboard",
  description: "Personal wealth intelligence platform for families",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={lora.variable}>
      <body className="paper-bg antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
