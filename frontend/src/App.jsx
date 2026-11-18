
import { useEffect, useState } from 'react'
import './App.css'

const API = '/api'
const ROLES = ['top', 'jungle', 'mid', 'bottom', 'support']

const emptyProfile = {
  displayName: '',
  username: '',
  preferredChampions: [],
  assignedRole: 'mid',
}

async function api(path, options = {}) {
  const response = await fetch(`${API}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`)
  }

  const text = await response.text()
  return text ? JSON.parse(text) : null
}

function App() {
  const [page, setPage] = useState('draft')
  const [champions, setChampions] = useState([])
  const [profile, setProfile] = useState(() => {
    try {
      return JSON.parse(localStorage.getItem('pb-profile')) || emptyProfile
    } catch {
      return emptyProfile
    }
  })

  const [side, setSide] = useState('blue')
  const [actions, setActions] = useState([])
  const [selectedSlot, setSelectedSlot] = useState(null)
  const [opponentRole, setOpponentRole] = useState('bottom')
  const [winRates, setWinRates] = useState({})
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function initialise() {
      try {
        const [champData, gameData] = await Promise.all([
          api('/champions'),
          api('/game'),
        ])

        setChampions(champData.champs || [])
        setActions(gameData.actions || [])

        if (gameData.side) {
          setSide(gameData.side)
        }
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    initialise()
  }, [])

  const opponentSide = side === 'blue' ? 'red' : 'blue'

  function getSlots(actionSide, type) {
    return actions.filter(
        action => action.side === actionSide && action.type === type
    )
  }

  async function saveProfile(event) {
    event.preventDefault()
    setError('')

    try {
      await api('/user', {
        method: 'POST',
        body: JSON.stringify(profile),
      })

      await api(`/side?side=${side}`, {
        method: 'POST',
      })

      localStorage.setItem('pb-profile', JSON.stringify(profile))
      setPage('draft')
    } catch (err) {
      setError(err.message)
    }
  }

  async function changeSide(newSide) {
    setSide(newSide)
    setError('')

    try {
      await api(`/side?side=${newSide}`, {
        method: 'POST',
      })
    } catch (err) {
      setError(err.message)
    }
  }

async function selectChampion(champion) {
  if (!selectedSlot) return

  const { actionSide, type } = selectedSlot
  const isOpponent = actionSide === opponentSide

  const action = {
    side: actionSide,
    type,
    champion: champion.toLowerCase(),
    role: isOpponent ? opponentRole : profile.assignedRole,
  }

  try {
    setError('')

    const result = await api('/action', {
      method: 'POST',
      body: JSON.stringify({ action }),
    })

    setActions(previous => [...previous, action])

    // API returns: {"Katarina": 51.75, "Xayah": 49.35}
    if (isOpponent && type === 'pick' && result) {
      setWinRates(result)
    }

    setSelectedSlot(null)
  } catch (err) {
    setError(err.message)
  }
}

function MatchupResults() {
  const results = Object.entries(winRates)
    .filter(([champion, rate]) => typeof rate === 'number')
    .sort((a, b) => b[1] - a[1])

  if (!results.length) {
    return (
      <div className="empty-results">
        Matchup statistics will appear when your opponent picks a champion.
      </div>
    )
  }

  return (
    <div className="results-grid">
      {results.map(([champion, rate], index) => (
        <div className="result-card" key={champion}>
          <div className="result-top">
            <strong>{champion}</strong>
            {index === 0 && (
              <span className="best-matchup">BEST MATCHUP</span>
            )}
          </div>

          <span className={`win-rate ${
            rate >= 50 ? 'positive' : 'negative'
          }`}>
            {rate.toFixed(2)}%
          </span>

          <div className="win-rate-track">
            <div
              className={`win-rate-fill ${
                rate >= 50 ? 'positive' : 'negative'
              }`}
              style={{ width: `${Math.min(100, Math.max(0, rate))}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  )
}

  function TeamPanel({ teamSide }) {
    const isUser = teamSide === side
    const bans = getSlots(teamSide, 'ban')
    const picks = getSlots(teamSide, 'pick')
    const slots = [
      { title: 'BANS', type: 'ban', count: 5, values: bans },
      { title: 'PICKS', type: 'pick', count: 5, values: picks },
    ]

    return (
        <section className={`team-panel ${teamSide}`}>
          <div className="team-heading">
            <div>
            <span className="eyebrow">
              {isUser ? 'YOUR TEAM' : 'OPPONENT'}
            </span>
              <h2>{teamSide.toUpperCase()} SIDE</h2>
            </div>
            <span className="team-indicator">
            {isUser ? 'YOU' : 'ENEMY'}
          </span>
          </div>

          {slots.map(group => (
              <div className="slot-group" key={group.type}>
                <h3>{group.title}</h3>

                <div className="champion-slots">
                  {Array.from({ length: group.count }, (_, index) => {
                    const action = group.values[index]

                    return (
                        <button
                            key={index}
                            className={`champion-slot ${action ? 'filled' : ''}`}
                            onClick={() => {
                              if (!action) {
                                setSelectedSlot({
                                  actionSide: teamSide,
                                  type: group.type,
                                })
                              }
                            }}
                            disabled={!!action}
                        >
                          {action ? (
                              <>
                        <span className="champion-name">
                          {champions.find(
                              c => c.toLowerCase() === action.champion
                          ) || action.champion}
                        </span>
                                <span className="slot-number">
                          {index + 1}
                        </span>
                              </>
                          ) : (
                              <>
                                <span className="plus">+</span>
                                <span>{group.type === 'ban' ? 'Ban' : 'Pick'}</span>
                              </>
                          )}
                        </button>
                    )
                  })}
                </div>
              </div>
          ))}
        </section>
    )
  }

  if (loading) {
    return <div className="loading">Loading PB&J...</div>
  }

  return (
      <div className="app">
        <header className="navbar">
          <button className="brand" onClick={() => setPage('draft')}>
            <span className="brand-icon">PB</span>
            <span>PB<span className="brand-accent">&</span>J</span>
          </button>

          <nav>
            <button
                className={page === 'draft' ? 'active' : ''}
                onClick={() => setPage('draft')}
            >
              Draft
            </button>
            <button
                className={page === 'profile' ? 'active' : ''}
                onClick={() => setPage('profile')}
            >
              Player settings
            </button>
          </nav>

          <span className="version">PICK / BAN HELPER</span>
        </header>

        {error && <div className="error">{error}</div>}

        {page === 'profile' ? (
            <main className="settings-page">
              <div className="page-heading">
                <span className="eyebrow">CONFIGURATION</span>
                <h1>Player details</h1>
                <p>Configure your preferences before entering champion select.</p>
              </div>

              <form className="settings-card" onSubmit={saveProfile}>
                <label>
                  Display name
                  <input
                      value={profile.displayName}
                      onChange={e => setProfile({
                        ...profile,
                        displayName: e.target.value,
                      })}
                      required
                  />
                </label>

                <label>
                  Username
                  <input
                      value={profile.username}
                      onChange={e => setProfile({
                        ...profile,
                        username: e.target.value,
                      })}
                      required
                  />
                </label>

                <label>
                  Assigned role
                  <select
                      value={profile.assignedRole}
                      onChange={e => setProfile({
                        ...profile,
                        assignedRole: e.target.value,
                      })}
                  >
                    {ROLES.map(role => (
                        <option key={role} value={role}>
                          {role.toUpperCase()}
                        </option>
                    ))}
                  </select>
                </label>

                <div className="field">
                  <label>Preferred champions</label>
                  <p className="field-description">
                    Select up to five champions from the available champion list.
                  </p>

                  <div className="champion-picker-list">
                    {champions.map(champion => {
                      const selected = profile.preferredChampions.includes(
                          champion.toLowerCase()
                      )

                      return (
                          <button
                              type="button"
                              key={champion}
                              className={`champion-choice ${selected ? 'selected' : ''}`}
                              onClick={() => {
                                const key = champion.toLowerCase()
                                setProfile(previous => ({
                                  ...previous,
                                  preferredChampions: selected
                                      ? previous.preferredChampions.filter(c => c !== key)
                                      : previous.preferredChampions.length < 5
                                          ? [...previous.preferredChampions, key]
                                          : previous.preferredChampions,
                                }))
                              }}
                          >
                            {selected && '✓ '}
                            {champion}
                          </button>
                      )
                    })}
                  </div>

                  <span className="field-description">
                {profile.preferredChampions.length}/5 selected
              </span>
                </div>

                <button className="primary-button" type="submit">
                  Save player details
                </button>
              </form>
            </main>
        ) : (
            <main className="draft-page">
              <div className="page-heading draft-heading">
                <div>
                  <span className="eyebrow">CHAMPION SELECT</span>
                  <h1>Draft assistant</h1>
                  <p>Track the draft and find favourable matchups.</p>
                </div>

                <button
                    className="secondary-button"
                    onClick={() => setPage('profile')}
                >
                  ⚙ Player settings
                </button>
              </div>

              <div className="draft-toolbar">
                <div>
                  <span className="toolbar-label">YOUR TEAM COLOUR</span>
                  <div className="side-selector">
                    <button
                        className={side === 'blue' ? 'blue active' : 'blue'}
                        onClick={() => changeSide('blue')}
                    >
                      Blue side
                    </button>
                    <button
                        className={side === 'red' ? 'red active' : 'red'}
                        onClick={() => changeSide('red')}
                    >
                      Red side
                    </button>
                  </div>
                </div>

                <div className="player-summary">
                  <span className="toolbar-label">PLAYER</span>
                  <strong>{profile.displayName || 'Not configured'}</strong>
                  <span>{profile.assignedRole.toUpperCase()}</span>
                </div>
              </div>

              <div className="draft-board">
                <TeamPanel teamSide={side} />
                <div className="board-divider">
                  <span>VS</span>
                </div>
                <TeamPanel teamSide={opponentSide} />
              </div>

              {side !== '' && (
                  <div className="opponent-role">
                    <label htmlFor="opponent-role">Opponent's selected role</label>
                    <select
                        id="opponent-role"
                        value={opponentRole}
                        onChange={e => setOpponentRole(e.target.value)}
                    >
                      {ROLES.map(role => (
                          <option key={role} value={role}>
                            {role.toUpperCase()}
                          </option>
                      ))}
                    </select>
                  </div>
              )}

              <section className="results-section">
                <div className="results-heading">
                  <span className="eyebrow">MATCHUP ANALYSIS</span>
                  <h2>Your champion matchups</h2>
                  <p>Win rates against the latest opposing pick.</p>
                </div>

                <MatchupResults />
              </section>
            </main>
        )}

        {selectedSlot && (
            <div
                className="modal-backdrop"
                onClick={() => setSelectedSlot(null)}
            >
              <div className="champion-modal" onClick={e => e.stopPropagation()}>
                <div className="modal-heading">
                  <div>
                    <span className="eyebrow">CHAMPION SELECT</span>
                    <h2>
                      Select a {selectedSlot.type}
                    </h2>
                  </div>
                  <button onClick={() => setSelectedSlot(null)}>✕</button>
                </div>

                <div className="champion-picker-list modal-list">
                  {champions.map(champion => (
                      <button
                          key={champion}
                          className="champion-choice"
                          onClick={() => selectChampion(champion)}
                      >
                        {champion}
                      </button>
                  ))}
                </div>
              </div>
            </div>
        )}
      </div>
  )
}

export default App
