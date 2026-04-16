import React from "react";
import "./globals.css";
import Navbar from "./components/Navbar";

export const metadata = {
  title: "InDel Portal",
  description: "AI-powered income protection for gig workers."
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Navbar />
        <div className="pt-16">{children}</div>
      </body>
    </html>
  );
}
