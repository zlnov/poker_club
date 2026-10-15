/**
 * Marketing landing page for Poker Club.
 *
 * Served at the application root (`/`). Demonstrates the product with
 * real UI screenshots and routes auth CTAs to existing `/login` and `/register`.
 */

import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMediaQuery } from '@mantine/hooks'
import {
  IconArrowRight,
  IconCalculator,
  IconChartBar,
  IconMenu2,
  IconPlayerPlay,
  IconUsers,
  IconX,
} from '@tabler/icons-react'
import { useAuth } from '../auth'
import styles from './landing/LandingPage.module.css'

const SECTIONS = [
  { id: 'features', label: 'Возможности' },
  { id: 'how-it-works', label: 'Как это работает' },
  { id: 'statistics', label: 'Статистика' },
] as const

const SCENARIO_STEPS = [
  {
    label: 'Создание игры',
    hint: 'Соберите игроков за столом',
    title: 'Соберите игроков за столом',
    text: 'Создайте игру в клубе, задайте параметры и добавьте участников. Организатор начинает работу с одной записи игры, а не с разрозненных сообщений и заметок.',
    image: '/landing/block3/1-create-game.png',
    imageWidth: 688,
    imageHeight: 1217,
  },
  {
    label: 'Ход игры',
    hint: 'Учёт под контролем',
    title: 'Игра идёт — учёт под контролем',
    text: 'Фиксируйте участников, бай-ины и ребаи в процессе игры. Информация сохраняется в приложении, что позволяет не восстанавливать суммы по памяти после завершения.',
    image: '/landing/block3/2-game-flow.png',
    imageWidth: 861,
    imageHeight: 879,
  },
  {
    label: 'Итоги игры',
    hint: 'Получите результат',
    title: 'Завершите игру — получите результат',
    text: 'После завершения игры фиксируются выплаты и рассчитываются финансовые результаты участников на основе сохранённых данных.',
    image: '/landing/block3/3-finish-game.png',
    imageWidth: 857,
    imageHeight: 1201,
  },
  {
    label: 'История и статистика',
    hint: 'Каждая игра пополняет историю',
    title: 'Каждая игра пополняет историю',
    text: 'Результаты завершённых игр становятся основой для истории клуба и накопительной статистики игроков.',
    image: '/landing/block3/4-game-statistic.png',
    imageWidth: 858,
    imageHeight: 1441,
  },
] as const

const SCENARIO_PREVIEW_SCALE = 4
const SCENARIO_SHOT_MAX_H_MOBILE = 360
const SCENARIO_SHOT_MAX_H_DESKTOP = 480
const SCENARIO_STEPS_WIDTH = 260
const SCENARIO_CARD_WIDTH = 280
const SCENARIO_GAP = 24

/** Fit screenshot into max height/width while preserving aspect ratio. */
function getScenarioShotSize(
  imageWidth: number,
  imageHeight: number,
  maxHeight: number,
  maxWidth: number,
): { width: number; height: number } {
  const aspect = imageWidth / Math.max(imageHeight, 1)
  let height = maxHeight
  let width = height * aspect

  if (width > maxWidth && maxWidth > 0) {
    width = maxWidth
    height = width / aspect
  }

  return {
    width: Math.round(width * 10) / 10,
    height: Math.round(height * 10) / 10,
  }
}

const STAT_VIEWS = [
  {
    id: 'players',
    tab: 'Игроки',
    title: 'Кто играет результативнее?',
    text: 'Смотрите историю выступлений игроков, их прибыль и другие доступные показатели за сыгранные игры.',
    image: '/landing/block4/1-members-stat.png',
  },
  {
    id: 'games',
    tab: 'Игры',
    title: 'Результаты не теряются',
    text: 'Возвращайтесь к завершённым играм и сохранённым итогам — история клуба всегда под рукой.',
    image: '/landing/block4/2-player-stat.png',
  },
  {
    id: 'club',
    tab: 'Клуб',
    title: 'Вся история клуба в одном месте',
    text: 'Сводные показатели активности клуба: завершённые игры, участники и динамика результатов.',
    image: '/landing/block4/3-club-stat.png',
  },
] as const

const PROBLEM_CARDS = [
  {
    icon: IconCalculator,
    title: 'Учёт без путаницы',
    problem:
      'Buy-in, Rebuy приходится учитывать вручную.',
    solution:
      'Удобный учёт участников и операций в процессе игры.',
  },
  {
    icon: IconChartBar,
    title: 'История результатов',
    problem:
      'Сложно отследить результаты прошлых игр.',
    solution:
      'История игр и автоматически обновляемая статистика игроков.',
  },
  {
    icon: IconUsers,
    title: 'Управление клубом',
    problem:
      'Информация о клубе разбросана в разных источника.',
    solution:
      'Единое пространство клуба для управления участниками и играми.',
  },
] as const

const HERO_SHOTS = [
  {
    src: '/landing/block1/sshot-dash.png',
    className: 'shotMain',
    label: 'Дашборд клуба',
    previewScale: 3,
  },
  {
    src: '/landing/block1/sshot-stat.png',
    className: 'shotTop',
    label: 'Статистика',
    previewScale: 5,
  },
  {
    src: '/landing/block1/sshot-games.png',
    className: 'shotLeft',
    label: 'Список игр',
    previewScale: 5,
  },
  {
    src: '/landing/block1/sshot-members.png',
    className: 'shotRight',
    label: 'Участники клуба',
    previewScale: 5,
  },
] as const
const IMAGE_FLIP_MS = 380
const IMAGE_FLIP_EASING = 'cubic-bezier(0.22, 1, 0.36, 1)'

interface OriginRect {
  top: number
  left: number
  width: number
  height: number
}

interface ImagePreview {
  src: string
  label: string
  width: number
  height: number
  origin: OriginRect
}

function scrollToSection(id: string) {
  const el = document.getElementById(id)
  if (!el) return
  el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/** FLIP invert: move/scale the final-sized image back onto the thumbnail. */
function getFlipTransform(origin: OriginRect, last: DOMRect): string {
  const originCx = origin.left + origin.width / 2
  const originCy = origin.top + origin.height / 2
  const lastCx = last.left + last.width / 2
  const lastCy = last.top + last.height / 2
  const dx = originCx - lastCx
  const dy = originCy - lastCy
  const sx = origin.width / Math.max(last.width, 1)
  const sy = origin.height / Math.max(last.height, 1)
  return `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})`
}

export function LandingPage() {
  const navigate = useNavigate()
  const { startDemo, isLoading: authLoading } = useAuth()
  const [menuOpen, setMenuOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  const [scenarioStep, setScenarioStep] = useState(0)
  const [statView, setStatView] = useState(0)
  const [imagePreview, setImagePreview] = useState<ImagePreview | null>(null)
  const [scenarioLayoutWidth, setScenarioLayoutWidth] = useState(0)
  const [demoStarting, setDemoStarting] = useState(false)
  const [demoError, setDemoError] = useState<string | null>(null)
  const lightboxRef = useRef<HTMLDivElement>(null)
  const lightboxImageRef = useRef<HTMLImageElement>(null)
  const scenarioLayoutRef = useRef<HTMLDivElement>(null)
  const imageClosingRef = useRef(false)
  const imagePreviewRef = useRef<ImagePreview | null>(null)
  const closeImagePreviewRef = useRef<() => void>(() => {})
  const isScenarioDesktop = useMediaQuery('(min-width: 960px)')

  imagePreviewRef.current = imagePreview

  const goToSection = (id: string) => {
    setMenuOpen(false)
    scrollToSection(id)
  }

  const handleStartDemo = async () => {
    setDemoError(null)
    setDemoStarting(true)
    try {
      const clubId = await startDemo()
      navigate(`/clubs/${clubId}`)
    } catch {
      setDemoError(
        'Демо сейчас недоступно. Попробуйте позже или создайте аккаунт.',
      )
    } finally {
      setDemoStarting(false)
    }
  }

  const openImagePreview = (
    src: string,
    label: string,
    previewScale: number,
    target: HTMLElement,
  ) => {
    if (imagePreviewRef.current) return

    const rect = target.getBoundingClientRect()
    const maxWidth = window.innerWidth * 0.92
    const maxHeight = window.innerHeight * 0.88
    const scaledWidth = rect.width * previewScale
    const scaledHeight = rect.height * previewScale
    const fit = Math.min(1, maxWidth / scaledWidth, maxHeight / scaledHeight)

    imageClosingRef.current = false
    setImagePreview({
      src,
      label,
      width: scaledWidth * fit,
      height: scaledHeight * fit,
      origin: {
        top: rect.top,
        left: rect.left,
        width: rect.width,
        height: rect.height,
      },
    })
  }

  const closeImagePreview = () => {
    const preview = imagePreviewRef.current
    if (!preview || imageClosingRef.current) return
    imageClosingRef.current = true

    const img = lightboxImageRef.current
    const backdrop = lightboxRef.current

    if (!img || prefersReducedMotion()) {
      setImagePreview(null)
      imageClosingRef.current = false
      return
    }

    const last = img.getBoundingClientRect()
    img.style.transition = `transform ${IMAGE_FLIP_MS}ms ${IMAGE_FLIP_EASING}`
    img.style.transform = getFlipTransform(preview.origin, last)
    if (backdrop) {
      backdrop.style.transition = `opacity ${Math.round(IMAGE_FLIP_MS * 0.85)}ms ease`
      backdrop.style.opacity = '0'
    }

    let finished = false
    const finish = () => {
      if (finished) return
      finished = true
      img.removeEventListener('transitionend', onEnd)
      setImagePreview(null)
      imageClosingRef.current = false
    }
    const onEnd = (event: TransitionEvent) => {
      if (event.target === img && event.propertyName === 'transform') finish()
    }

    img.addEventListener('transitionend', onEnd)
    window.setTimeout(finish, IMAGE_FLIP_MS + 80)
  }

  closeImagePreviewRef.current = closeImagePreview

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    const el = scenarioLayoutRef.current
    if (!el) return

    const updateWidth = () => setScenarioLayoutWidth(el.clientWidth)
    updateWidth()

    const observer = new ResizeObserver(updateWidth)
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    if (!imagePreview) return

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') closeImagePreviewRef.current()
    }

    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', onKeyDown)

    return () => {
      document.body.style.overflow = previousOverflow
      window.removeEventListener('keydown', onKeyDown)
    }
  }, [imagePreview])

  // Shared-element open: invert from thumbnail, then animate to final size.
  useLayoutEffect(() => {
    if (!imagePreview || imageClosingRef.current) return

    const img = lightboxImageRef.current
    const backdrop = lightboxRef.current
    if (!img) return

    if (prefersReducedMotion()) {
      if (backdrop) backdrop.style.opacity = '1'
      img.style.transform = 'none'
      return
    }

    const last = img.getBoundingClientRect()
    img.style.transition = 'none'
    img.style.transform = getFlipTransform(imagePreview.origin, last)
    if (backdrop) {
      backdrop.style.transition = 'none'
      backdrop.style.opacity = '0'
    }

    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        img.style.transition = `transform ${IMAGE_FLIP_MS}ms ${IMAGE_FLIP_EASING}`
        img.style.transform = 'none'
        if (backdrop) {
          backdrop.style.transition = `opacity ${Math.round(IMAGE_FLIP_MS * 0.85)}ms ease`
          backdrop.style.opacity = '1'
        }
      })
    })
  }, [imagePreview])

  const activeScenario = SCENARIO_STEPS[scenarioStep]
  const activeStat = STAT_VIEWS[statView]

  const scenarioShotMaxH = isScenarioDesktop
    ? SCENARIO_SHOT_MAX_H_DESKTOP
    : SCENARIO_SHOT_MAX_H_MOBILE

  let scenarioShotMaxW =
    scenarioShotMaxH * (activeScenario.imageWidth / activeScenario.imageHeight)
  if (scenarioLayoutWidth > 0) {
    scenarioShotMaxW = isScenarioDesktop
      ? Math.max(
          120,
          scenarioLayoutWidth -
            SCENARIO_STEPS_WIDTH -
            SCENARIO_CARD_WIDTH -
            SCENARIO_GAP * 2,
        )
      : scenarioLayoutWidth
  }

  const scenarioShotSize = getScenarioShotSize(
    activeScenario.imageWidth,
    activeScenario.imageHeight,
    scenarioShotMaxH,
    scenarioShotMaxW,
  )

  return (
    <div className={styles.page}>
      <header
        className={`${styles.header} ${scrolled ? styles.headerScrolled : ''}`}
      >
        <div className={styles.headerInner}>
          <a
            href="#top"
            className={styles.brand}
            onClick={(e) => {
              e.preventDefault()
              window.scrollTo({ top: 0, behavior: 'smooth' })
            }}
          >
            <img
              src="/landing/logo/logo.png"
              alt=""
              className={styles.brandLogo}
              width={36}
              height={36}
            />
            <span className={styles.brandName}>Poker Club</span>
          </a>

          <nav className={styles.nav} aria-label="Разделы страницы">
            {SECTIONS.map((section) => (
              <button
                key={section.id}
                type="button"
                className={styles.navLink}
                onClick={() => goToSection(section.id)}
              >
                {section.label}
              </button>
            ))}
          </nav>

          <div className={styles.headerActions}>
            <Link to="/login" className={styles.btnLogin}>
              Войти
            </Link>
            <Link to="/register" className={styles.btnRegister}>
              Регистрация
            </Link>
          </div>

          <button
            type="button"
            className={styles.mobileMenuBtn}
            aria-label={menuOpen ? 'Закрыть меню' : 'Открыть меню'}
            aria-expanded={menuOpen}
            onClick={() => setMenuOpen((open) => !open)}
          >
            {menuOpen ? <IconX size={20} /> : <IconMenu2 size={20} />}
          </button>
        </div>

        {menuOpen ? (
          <div className={styles.mobileNav}>
            {SECTIONS.map((section) => (
              <button
                key={section.id}
                type="button"
                className={styles.navLink}
                onClick={() => goToSection(section.id)}
              >
                {section.label}
              </button>
            ))}
            <div className={styles.mobileNavActions}>
              <Link
                to="/login"
                className={styles.btnSecondary}
                onClick={() => setMenuOpen(false)}
              >
                Войти
              </Link>
              <Link
                to="/register"
                className={styles.btnPrimary}
                onClick={() => setMenuOpen(false)}
              >
                Регистрация
              </Link>
            </div>
          </div>
        ) : null}
      </header>

      <main id="top">
        {/* Block 01 — Hero */}
        <section className={styles.hero} aria-labelledby="hero-title">
          <div className={styles.heroInner}>
            <div className={styles.heroCopy}>
              <span className={styles.eyebrow}>Организация покерного клуба</span>
              <h1 id="hero-title" className={styles.heroTitle}>
                Ваш клуб. Ваши игры.{' '}
                <span className={styles.heroTitleAccent}>
                  Вся статистика в одном месте
                </span>
              </h1>
              <p className={styles.heroLead}>
                Poker Club помогает организовывать покерные игры, вести учёт
                результатов и сохранять историю клуба.
              </p>
              <div className={styles.heroActions}>
                <Link to="/register" className={styles.btnPrimary}>
                  Создать аккаунт
                  <IconArrowRight size={18} />
                </Link>
                <button
                  type="button"
                  className={styles.btnSecondary}
                  onClick={() => {
                    void handleStartDemo()
                  }}
                  disabled={demoStarting || authLoading}
                >
                  <IconPlayerPlay size={18} />
                  {demoStarting
                    ? 'Загрузка демо…'
                    : 'Посмотреть, как работает'}
                </button>
              </div>
              {demoError ? (
                <p className={styles.heroLead} role="alert">
                  {demoError}
                </p>
              ) : null}
            </div>

            <div className={styles.heroVisual}>
              <div className={styles.collage}>
                {HERO_SHOTS.map((shot) => {
                  const shotClass =
                    shot.className === 'shotMain'
                      ? styles.shotMain
                      : shot.className === 'shotTop'
                        ? styles.shotTop
                        : shot.className === 'shotLeft'
                          ? styles.shotLeft
                          : styles.shotRight

                  const isPreviewSource = imagePreview?.src === shot.src

                  return (
                    <button
                      key={shot.src}
                      type="button"
                      className={`${styles.collageShot} ${shotClass} ${
                        isPreviewSource ? styles.collageShotHidden : ''
                      }`}
                      aria-label={`Увеличить скриншот: ${shot.label}`}
                      onClick={(event) =>
                        openImagePreview(
                          shot.src,
                          shot.label,
                          shot.previewScale,
                          event.currentTarget,
                        )
                      }
                    >
                      <img src={shot.src} alt={shot.label} />
                    </button>
                  )
                })}
              </div>
            </div>
          </div>
        </section>

        {imagePreview ? (
          <div
            ref={lightboxRef}
            className={styles.heroLightbox}
            role="dialog"
            aria-modal="true"
            aria-label={imagePreview.label}
            onClick={closeImagePreview}
          >
            <button
              type="button"
              className={styles.heroLightboxClose}
              aria-label="Закрыть просмотр"
              onClick={(event) => {
                event.stopPropagation()
                closeImagePreview()
              }}
            >
              <IconX size={22} />
            </button>
            <img
              ref={lightboxImageRef}
              src={imagePreview.src}
              alt={imagePreview.label}
              className={styles.heroLightboxImage}
              style={{
                width: imagePreview.width,
                height: imagePreview.height,
              }}
              onClick={(event) => event.stopPropagation()}
            />
          </div>
        ) : null}

        {/* Block 02 — Problem → solution */}
        <section
          id="features"
          className={styles.section}
          aria-labelledby="features-title"
        >
          <div className={styles.sectionInner}>
            <div className={styles.sectionHeader}>
              <span className={styles.eyebrow}>Зачем нужен Poker Club</span>
              <h2 id="features-title" className={styles.sectionTitle}>
                Меньше ручной работы. Больше порядка в клубе
              </h2>
              <p className={styles.sectionLead}>
                Всё необходимое для организации игр и учёта результатов — в одном
                месте.
              </p>
            </div>

            <div className={styles.cards}>
              {PROBLEM_CARDS.map((card) => (
                <article key={card.title} className={styles.card}>
                  <div className={styles.cardIcon}>
                    <card.icon size={22} stroke={1.6} />
                  </div>
                  <h3 className={styles.cardTitle}>{card.title}</h3>
                  <p className={styles.cardLabel}>Проблема</p>
                  <p className={styles.cardText}>{card.problem}</p>
                  <hr className={styles.cardDivider} />
                  <p
                    className={`${styles.cardLabel} ${styles.cardLabelSolution}`}
                  >
                    Решение
                  </p>
                  <p className={styles.cardText}>{card.solution}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        {/* Block 03 — Game scenario */}
        <section
          id="how-it-works"
          className={`${styles.section} ${styles.scenario}`}
          aria-labelledby="scenario-title"
        >
          <div className={styles.sectionInner}>
            <div className={styles.sectionHeader}>
              <span className={styles.eyebrow}>Как это работает</span>
              <h2 id="scenario-title" className={styles.sectionTitle}>
                От первой раздачи до итоговой статистики
              </h2>
              <p className={styles.sectionLead}>
                Весь процесс организации игры — от создания до анализа
                результатов — в одном приложении.
              </p>
            </div>

            <div ref={scenarioLayoutRef} className={styles.scenarioLayout}>
              <div
                className={styles.steps}
                role="tablist"
                aria-label="Этапы игры"
              >
                {SCENARIO_STEPS.map((step, index) => (
                  <button
                    key={step.label}
                    type="button"
                    role="tab"
                    aria-selected={scenarioStep === index}
                    className={`${styles.stepBtn} ${
                      scenarioStep === index ? styles.stepActive : ''
                    }`}
                    onClick={() => setScenarioStep(index)}
                  >
                    <span className={styles.stepNum}>{index + 1}</span>
                    <span>
                      <span className={styles.stepLabel}>{step.label}</span>
                      <span className={styles.stepHint}>{step.hint}</span>
                    </span>
                  </button>
                ))}
              </div>

              <button
                type="button"
                className={`${styles.scenarioShotWrap} ${
                  imagePreview?.src === activeScenario.image
                    ? styles.scenarioShotHidden
                    : ''
                }`}
                style={{
                  width: scenarioShotSize.width,
                  height: scenarioShotSize.height,
                }}
                aria-label={`Увеличить скриншот: ${activeScenario.label}`}
                onClick={(event) =>
                  openImagePreview(
                    activeScenario.image,
                    activeScenario.label,
                    SCENARIO_PREVIEW_SCALE,
                    event.currentTarget,
                  )
                }
              >
                <img
                  src={activeScenario.image}
                  alt={`Экран приложения: ${activeScenario.label}`}
                  className={styles.scenarioShot}
                  width={activeScenario.imageWidth}
                  height={activeScenario.imageHeight}
                />
              </button>

              <div className={styles.scenarioCard}>
                <div>
                  <div className={styles.scenarioProgress}>
                    {scenarioStep + 1} / {SCENARIO_STEPS.length}
                  </div>
                  <div className={styles.progressBar} aria-hidden="true">
                    {SCENARIO_STEPS.map((_, index) => (
                      <span
                        key={SCENARIO_STEPS[index].label}
                        className={`${styles.progressSeg} ${
                          index <= scenarioStep ? styles.progressSegOn : ''
                        }`}
                      />
                    ))}
                  </div>
                </div>
                <h3 className={styles.scenarioCardTitle}>
                  {activeScenario.title}
                </h3>
                <p className={styles.scenarioCardTag}>{activeScenario.label}</p>
                <p className={styles.scenarioCardText}>{activeScenario.text}</p>
                <button
                  type="button"
                  className={`${styles.btnSecondary} ${styles.scenarioNext}`}
                  onClick={() =>
                    setScenarioStep((step) => (step + 1) % SCENARIO_STEPS.length)
                  }
                >
                  Далее
                  <IconArrowRight size={16} />
                </button>
              </div>
            </div>
          </div>
        </section>

        {/* Block 04 — Statistics */}
        <section
          id="statistics"
          className={styles.section}
          aria-labelledby="stats-title"
        >
          <div className={styles.sectionInner}>
            <div className={styles.sectionHeader}>
              <span className={styles.eyebrow}>Статистика</span>
              <h2 id="stats-title" className={styles.sectionTitle}>
                Цифры, которые рассказывают историю клуба
              </h2>
              <p className={styles.sectionLead}>
                Следите за результатами отдельных игроков и динамикой клуба на
                основе сыгранных игр.
              </p>
            </div>

            <div
              className={styles.statsTabs}
              role="tablist"
              aria-label="Представления статистики"
            >
              {STAT_VIEWS.map((view, index) => (
                <button
                  key={view.id}
                  type="button"
                  role="tab"
                  aria-selected={statView === index}
                  className={`${styles.statsTab} ${
                    statView === index ? styles.statsTabActive : ''
                  }`}
                  onClick={() => setStatView(index)}
                >
                  {view.tab}
                </button>
              ))}
            </div>

            <div className={styles.panelFade} key={activeStat.id}>
              <div className={styles.statsStage}>
                <img
                  src={activeStat.image}
                  alt={`Статистика: ${activeStat.tab}`}
                  className={styles.statsShot}
                />
              </div>
              <div className={styles.statsCaption}>
                <h3 className={styles.statsCaptionTitle}>{activeStat.title}</h3>
                <p className={styles.statsCaptionText}>{activeStat.text}</p>
              </div>
            </div>
          </div>
        </section>

        {/* Block 05 — CTA */}
        <section id="cta" className={styles.cta} aria-labelledby="cta-title">
          <h2 id="cta-title" className={styles.ctaTitle}>
            Ваш клуб. Ваша история.
          </h2>
          <p className={styles.ctaLead}>
            Начните вести игры и сохранять результаты клуба в одном месте.
          </p>
          <div className={styles.ctaActions}>
            <Link to="/register" className={styles.btnPrimary}>
              Создать аккаунт
            </Link>
            <Link to="/login" className={styles.btnGhost}>
              У меня уже есть аккаунт
            </Link>
          </div>
        </section>
      </main>

      <footer className={styles.footer}>
        <div className={styles.footerInner}>
          <div className={styles.footerGrid}>
            <div>
              <p className={styles.footerBrandName}>Poker Club</p>
              <p className={styles.footerBrandText}>
                Инструмент для организации покерного клуба и учёта результатов
                игр.
              </p>
            </div>

            <div className={styles.footerCols}>
              <div>
                <p className={styles.footerColTitle}>Продукт</p>
                <ul className={styles.footerLinks}>
                  {SECTIONS.map((section) => (
                    <li key={section.id}>
                      <button
                        type="button"
                        className={styles.footerLink}
                        onClick={() => goToSection(section.id)}
                      >
                        {section.label}
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <p className={styles.footerColTitle}>Аккаунт</p>
                <ul className={styles.footerLinks}>
                  <li>
                    <Link to="/register" className={styles.footerLink}>
                      Регистрация
                    </Link>
                  </li>
                  <li>
                    <Link to="/login" className={styles.footerLink}>
                      Войти
                    </Link>
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <div className={styles.footerBottom}>
            <span>© 2026 Poker Club</span>
          </div>
        </div>
      </footer>
    </div>
  )
}
