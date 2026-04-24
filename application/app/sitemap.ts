import type { MetadataRoute } from 'next'
import { getAllStages } from '@/lib/stages'

const siteUrl = 'https://go-backend-production.vercel.app'

export default function sitemap(): MetadataRoute.Sitemap {
  const stages = getAllStages()

  const stageUrls = stages.map((s) => ({
    url: `${siteUrl}/stages/${s.slug}`,
    lastModified: new Date(),
  }))

  return [
    {
      url: siteUrl,
      lastModified: new Date(),
    },
    ...stageUrls,
  ]
}
