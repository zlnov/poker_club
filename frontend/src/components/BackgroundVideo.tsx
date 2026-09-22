/**
 * BackgroundVideo component.
 *
 * Renders a background video with a static fallback image.
 * The video is decorative and does not block UI interaction.
 *
 * Requirements:
 * - Video and React app start simultaneously; UI doesn't wait for video
 * - UI remains fully clickable during video playback
 * - Video works asynchronously from React app
 * - Uses autoplay, muted, playsInline
 * - Does NOT use loop
 * - After 'ended', video hides and static poster remains
 * - Static poster is below video and used as fallback
 * - Video doesn't restart on navigation between pages
 * - Different video/image for desktop and mobile
 * - Uses fixed files from /public/background/
 * - No code changes needed to swap video/image files
 * - Works regardless of video duration
 * - If video fails to load or autoplay fails, app continues with static background
 */

import { useEffect, useRef, useState } from 'react'
import { useMediaQuery } from '@mantine/hooks'

interface BackgroundVideoProps {
  /** Additional class name for the container */
  className?: string
}

// Fixed file paths - changing these files doesn't require code changes
const DESKTOP_VIDEO = '/background/background-desktop.mp4'
const MOBILE_VIDEO = '/background/background-mobile.mp4'
const DESKTOP_POSTER = '/background/background-desktop.jpg'
const MOBILE_POSTER = '/background/background-mobile.jpg'

export function BackgroundVideo({ className }: BackgroundVideoProps) {
  const isMobile = useMediaQuery('(max-width: 768px)')
  const videoRef = useRef<HTMLVideoElement>(null)
  const [hasEnded, setHasEnded] = useState(false)
  const [hasError, setHasError] = useState(false)

  const videoSrc = isMobile ? MOBILE_VIDEO : DESKTOP_VIDEO
  const posterSrc = isMobile ? MOBILE_POSTER : DESKTOP_POSTER

  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    // Reset state when video source changes (desktop/mobile switch)
    setHasEnded(false)
    setHasError(false)

    const handleEnded = () => {
      setHasEnded(true)
    }

    const handleError = () => {
      setHasError(true)
    }

    video.addEventListener('ended', handleEnded)
    video.addEventListener('error', handleError)

    return () => {
      video.removeEventListener('ended', handleEnded)
      video.removeEventListener('error', handleError)
    }
  }, [videoSrc])

  // If video has ended or errored, only show the poster
  const showVideo = !hasEnded && !hasError

  return (
    <div
      className={className}
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        zIndex: 0,
        overflow: 'hidden',
        pointerEvents: 'none',
      }}
    >
      {/* Static poster (fallback and background) */}
      <img
        src={posterSrc}
        alt=""
        aria-hidden="true"
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          objectFit: 'cover',
          zIndex: 1,
        }}
      />

      {/* Video layer (above poster) */}
      {showVideo && (
        <video
          ref={videoRef}
          src={videoSrc}
          autoPlay
          muted
          playsInline
          // No loop - video plays once
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            zIndex: 2,
          }}
          // Ensure video doesn't capture pointer events - UI remains clickable
          onMouseDown={(e) => e.preventDefault()}
        />
      )}
    </div>
  )
}
