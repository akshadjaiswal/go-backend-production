import type { Metadata, Viewport } from 'next'
import { Playfair_Display, Source_Serif_4, JetBrains_Mono } from 'next/font/google'
import './globals.css'
import { Sidebar } from '@/components/sidebar'
import { Header } from '@/components/header'
import { ThemeProvider } from '@/components/theme-provider'
import { getGitHubStars } from '@/lib/github'

const playfair = Playfair_Display({
  subsets: ['latin'],
  variable: '--font-heading',
  display: 'swap',
})

const sourceSerif = Source_Serif_4({
  subsets: ['latin'],
  variable: '--font-body',
  display: 'swap',
})

const jetbrains = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  display: 'swap',
})

const siteUrl = 'https://go-backend-production.vercel.app'

export const metadata: Metadata = {
  title: {
    default: 'Go Backend Production',
    template: '%s — Go Backend Production',
  },
  description:
    'A 10-stage Go backend learning resource — from HTTP basics to Docker deployment. Built with real production patterns.',
  icons: {
    icon: '/icon.svg',
  },
  metadataBase: new URL(siteUrl),
  openGraph: {
    title: 'Go Backend Production',
    description:
      '10 stages of Go backend development — HTTP, routing, auth, database, testing, and Docker deployment.',
    url: siteUrl,
    siteName: 'Go Backend Production',
    type: 'website',
    locale: 'en_US',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Go Backend Production',
    description:
      '10 stages of Go backend development — HTTP, routing, auth, database, testing, and Docker deployment.',
  },
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#000000',
}

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const stars = await getGitHubStars()

  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${playfair.variable} ${sourceSerif.variable} ${jetbrains.variable}`}
    >
      <head>
        {/* Blocking script: reads gbp_theme from localStorage before first paint to avoid FOUC */}
        <script
          dangerouslySetInnerHTML={{
            __html: `(function(){try{var t=localStorage.getItem('gbp_theme');if(t==='dark'||((!t||t==='system')&&window.matchMedia('(prefers-color-scheme: dark)').matches)){document.documentElement.classList.add('dark');}}catch(e){}})();`,
          }}
        />
      </head>
      <body className="font-body bg-background text-foreground antialiased">
        <ThemeProvider>
          <div className="flex min-h-screen">
            <Sidebar />
            <main className="flex-1 min-w-0">
              <Header stars={stars} />
              {children}
            </main>
          </div>
        </ThemeProvider>
      </body>
    </html>
  )
}
