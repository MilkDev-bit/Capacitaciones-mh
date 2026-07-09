export interface PresetAvatar {
  id: string
  name: string
  url: string
}

const svgToDataUrl = (svgString: string): string => {
  const cleaned = svgString
    .replace(/\n/g, '')
    .replace(/\s+/g, ' ')
    .trim()
  return `data:image/svg+xml;utf8,${encodeURIComponent(cleaned)}`
}

export const PRESET_AVATARS: PresetAvatar[] = [
  {
    id: 'avatar_graduado',
    name: 'Estudiante Graduado',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g1" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#6366F1"/>
          <stop offset="100%" stop-color="#8B5CF6"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g1)"/>
      <path d="M50 28 L18 42 L50 56 L82 42 Z" fill="#FFFFFF"/>
      <path d="M32 50 V66 C32 73 68 73 68 66 V50" fill="none" stroke="#FFFFFF" stroke-width="5" stroke-linecap="round"/>
      <path d="M78 44 V68" stroke="#FBBF24" stroke-width="4" stroke-linecap="round"/>
      <circle cx="78" cy="71" r="4" fill="#FBBF24"/>
    </svg>`)
  },
  {
    id: 'avatar_innovador',
    name: 'Mente Creativa',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g2" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#F59E0B"/>
          <stop offset="100%" stop-color="#EA580C"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g2)"/>
      <path d="M50 24 C38 24 29 33 29 45 C29 53 35 59 39 63 V71 C39 73 41 75 43 75 H57 C59 75 61 73 61 71 V63 C65 59 71 53 71 45 C71 33 62 24 50 24 Z" fill="#FFFFFF"/>
      <path d="M43 78 H57 M45 83 H55" stroke="#FFFFFF" stroke-width="4" stroke-linecap="round"/>
      <path d="M50 14 V18 M25 25 L28 28 M75 25 L72 28" stroke="#FFFFFF" stroke-width="3.5" stroke-linecap="round"/>
    </svg>`)
  },
  {
    id: 'avatar_developer',
    name: 'Desarrollador Tech',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g3" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#06B6D4"/>
          <stop offset="100%" stop-color="#3B82F6"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g3)"/>
      <path d="M36 34 L20 50 L36 66 M64 34 L80 50 L64 66" fill="none" stroke="#FFFFFF" stroke-width="7" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M56 28 L44 72" stroke="#FFFFFF" stroke-width="6" stroke-linecap="round"/>
    </svg>`)
  },
  {
    id: 'avatar_cohete',
    name: 'Pionero Espacial',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g4" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#F43F5E"/>
          <stop offset="100%" stop-color="#EC4899"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g4)"/>
      <path d="M50 20 C50 20 68 35 66 58 L50 52 L34 58 C32 35 50 20 50 20 Z" fill="#FFFFFF"/>
      <circle cx="50" cy="42" r="6" fill="#F43F5E"/>
      <path d="M34 58 L24 68 L36 65 M66 58 L76 68 L64 65" fill="#FFFFFF"/>
      <path d="M44 60 C44 70 50 82 50 82 C50 82 56 70 56 60 Z" fill="#FDE047"/>
    </svg>`)
  },
  {
    id: 'avatar_campeon',
    name: 'Campeón de Éxito',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g5" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#10B981"/>
          <stop offset="100%" stop-color="#0D9488"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g5)"/>
      <path d="M34 28 H66 V46 C66 55 59 62 50 62 C41 62 34 55 34 46 V28 Z" fill="#FFFFFF"/>
      <path d="M34 33 H25 C22 33 20 36 21 40 C22 45 26 49 34 50 M66 33 H75 C78 33 80 36 79 40 C78 45 74 49 66 50" fill="none" stroke="#FFFFFF" stroke-width="5" stroke-linecap="round"/>
      <path d="M50 62 V73 M36 76 H64" stroke="#FFFFFF" stroke-width="6" stroke-linecap="round"/>
      <circle cx="50" cy="43" r="5" fill="#10B981"/>
    </svg>`)
  },
  {
    id: 'avatar_mentor',
    name: 'Mentor Académico',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g6" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#8B5CF6"/>
          <stop offset="100%" stop-color="#D946EF"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g6)"/>
      <path d="M22 42 C32 40 44 43 50 47 C56 43 68 40 78 42 V72 C68 69 56 72 50 76 C44 72 32 69 22 72 V42 Z" fill="#FFFFFF"/>
      <path d="M50 47 V76" stroke="#8B5CF6" stroke-width="3" stroke-linecap="round"/>
      <path d="M50 25 L53 32 L60 33 L55 38 L56 45 L50 42 L44 45 L45 38 L40 33 L47 32 Z" fill="#FDE047"/>
    </svg>`)
  },
  {
    id: 'avatar_buho',
    name: 'Búho Sabio',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g7" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#475569"/>
          <stop offset="100%" stop-color="#4F46E5"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g7)"/>
      <path d="M26 30 L34 44 C34 44 40 38 50 38 C60 38 66 44 66 44 L74 30 L74 72 C74 78 64 82 50 82 C36 82 26 78 26 72 Z" fill="#FFFFFF"/>
      <circle cx="40" cy="53" r="8" fill="#475569"/>
      <circle cx="60" cy="53" r="8" fill="#475569"/>
      <circle cx="40" cy="53" r="3" fill="#FFFFFF"/>
      <circle cx="60" cy="53" r="3" fill="#FFFFFF"/>
      <path d="M50 60 L45 68 H55 Z" fill="#F59E0B"/>
    </svg>`)
  },
  {
    id: 'avatar_estrella',
    name: 'Estrella Destacada',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g8" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#EAB308"/>
          <stop offset="100%" stop-color="#EF4444"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g8)"/>
      <path d="M50 20 L58 39 L78 41 L63 54 L67 74 L50 64 L33 74 L37 54 L22 41 L42 39 Z" fill="#FFFFFF"/>
      <circle cx="50" cy="48" r="10" fill="#F59E0B"/>
    </svg>`)
  },
  {
    id: 'avatar_explorador',
    name: 'Brújula Exploradora',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g9" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#14B8A6"/>
          <stop offset="100%" stop-color="#2563EB"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g9)"/>
      <circle cx="50" cy="50" r="28" fill="none" stroke="#FFFFFF" stroke-width="5"/>
      <path d="M58 34 L53 53 L34 58 L47 47 Z" fill="#FFFFFF"/>
      <path d="M42 66 L47 47 L66 42 L53 53 Z" fill="#FDE047"/>
      <circle cx="50" cy="50" r="4" fill="#1E3A8A"/>
    </svg>`)
  },
  {
    id: 'avatar_creativo',
    name: 'Diseñador Creativo',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g10" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#A855F7"/>
          <stop offset="100%" stop-color="#EC4899"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g10)"/>
      <path d="M50 22 C34 22 22 34 22 50 C22 66 34 78 50 78 C56 78 60 74 60 68 C60 65 58 63 58 60 C58 56 61 54 65 54 H70 C75 54 78 50 78 44 C78 31 65 22 50 22 Z" fill="#FFFFFF"/>
      <circle cx="38" cy="42" r="5" fill="#EC4899"/>
      <circle cx="50" cy="35" r="5" fill="#A855F7"/>
      <circle cx="62" cy="42" r="5" fill="#3B82F6"/>
      <circle cx="38" cy="56" r="5" fill="#EAB308"/>
    </svg>`)
  },
  {
    id: 'avatar_estratega',
    name: 'Estratega Corona',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g11" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#3B82F6"/>
          <stop offset="100%" stop-color="#4F46E5"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g11)"/>
      <path d="M24 66 L20 38 L36 48 L50 30 L64 48 L80 38 L76 66 Z" fill="#FFFFFF"/>
      <rect x="26" y="70" width="48" height="6" rx="3" fill="#FDE047"/>
      <circle cx="50" cy="25" r="4" fill="#FDE047"/>
      <circle cx="20" cy="33" r="4" fill="#FDE047"/>
      <circle cx="80" cy="33" r="4" fill="#FDE047"/>
    </svg>`)
  },
  {
    id: 'avatar_guardian',
    name: 'Guardián de Honor',
    url: svgToDataUrl(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <linearGradient id="g12" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#EF4444"/>
          <stop offset="100%" stop-color="#E11D48"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="50" fill="url(#g12)"/>
      <path d="M50 24 L26 33 V52 C26 67 36 78 50 82 C64 78 74 67 74 52 V33 L50 24 Z" fill="#FFFFFF"/>
      <path d="M42 53 L48 59 L60 45" fill="none" stroke="#E11D48" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>`)
  }
]

/**
 * Obtiene el hash numérico de un string
 */
function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = (hash << 5) - hash + str.charCodeAt(i)
    hash |= 0
  }
  return Math.abs(hash)
}

/**
 * Retorna un avatar preestablecido de forma determinista usando el ID o nombre del usuario.
 */
export function getDefaultAvatarUrl(seed?: string | null): string {
  if (!seed || !seed.trim() || PRESET_AVATARS.length === 0) {
    return PRESET_AVATARS[0]?.url || ''
  }
  const index = hashString(seed.trim()) % PRESET_AVATARS.length
  return PRESET_AVATARS[index]?.url || PRESET_AVATARS[0]?.url || ''
}

/**
 * Si avatarUrl existe y no está vacío, lo devuelve.
 * Si no, asigna de forma automática y determinista una imagen/ícono preestablecido de alta calidad.
 */
export function getAvatarUrl(avatarUrl?: string | null, seed?: string | null): string {
  if (avatarUrl && avatarUrl.trim() !== '') {
    return avatarUrl
  }
  return getDefaultAvatarUrl(seed)
}
