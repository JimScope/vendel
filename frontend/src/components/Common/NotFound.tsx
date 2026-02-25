import { Link } from "@tanstack/react-router"
import { Button } from "@/components/ui/button"

const NotFound = () => {
  return (
    <div
      className="flex min-h-screen items-center justify-center flex-col p-4"
      data-testid="not-found"
    >
      <div className="flex items-center">
        <div className="flex flex-col ml-4 items-center justify-center p-4">
          <span className="text-6xl md:text-8xl font-serif leading-none mb-4">
            404
          </span>
          <span className="text-2xl font-serif mb-2">Oops!</span>
        </div>
      </div>

      <p className="text-lg text-muted-foreground mb-4 text-center">
        The page you are looking for was not found.
      </p>
      <Link to="/">
        <Button className="mt-4">Go Back</Button>
      </Link>
    </div>
  )
}

export default NotFound
