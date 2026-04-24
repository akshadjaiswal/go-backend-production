import { getGroups } from '@/lib/stages'
import { SidebarClient } from './sidebar-client'

export function Sidebar() {
  const groups = getGroups()
  return <SidebarClient groups={groups} />
}
