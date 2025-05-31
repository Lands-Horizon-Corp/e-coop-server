import { createFileRoute } from '@tanstack/react-router'
import { Button } from '@/components/ui/button' // Ensure the correct path and casing

export const Route = createFileRoute('/')({
  component: Index,
})

function Index() {
  return (
    <div className="p-4">
      <Button>Hello</Button>
    </div>
  )
}
