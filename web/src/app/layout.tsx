import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Bili-Up Web - Bilibili 视频管理平台",
  description: "一个用于管理 Bilibili 视频上传和字幕处理的 Web 平台",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body className="min-h-screen bg-gray-50">
        {children}
      </body>
    </html>
  );
}