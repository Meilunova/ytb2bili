/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    BACKEND_URL: process.env.BACKEND_URL || 'http://localhost:8096',
  },
  
  // API代理配置
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.BACKEND_URL || 'http://localhost:8096'}/api/:path*`,
      },
    ]
  },

  // 图片域名白名单配置
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'passport.bilibili.com',
        pathname: '/x/passport-tv-login/h5/qrcode/auth/**',
      },
    ],
  },

  // 生产模式下的静态导出配置（构建时可以取消注释）
  // output: 'export',
  // trailingSlash: true,
  // images: {
  //   unoptimized: true
  // },
  // distDir: 'out'
}

module.exports = nextConfig