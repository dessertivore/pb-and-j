
import { useEffect, useState } from 'react'
import './App.css'

const API = '/api'
const ROLES = ['top', 'jungle', 'mid', 'bottom', 'support']


const createEmptyProfiles = () =>
  Object.fromEntries(
    ROLES.map(role => [
      role,
      {
        displayName: '',
        username: '',
        assignedRole: role,
        preferredChampions: [],
      },
    ])
  )

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
  const [activeRole, setActiveRole] = useState('top')
  const [page, setPage] = useState('draft')
  const [pickLoading, setPickLoading] = useState(false)
  const [allChampions, setAllChampions] = useState([])
  const [availableChampions, setAvailableChampions] = useState([])

  const [profiles, setProfiles] = useState(createEmptyProfiles)

  const [side, setSide] = useState('blue')
  const [actions, setActions] = useState([])
  const [selectedSlot, setSelectedSlot] = useState(null)  
  const [winRates, setWinRates] = useState({})

  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function initialise() {
      try {
        const [champData, gameData, userData] = await Promise.all([
          api('/champions?all=true'),
          api('/game'),
          api('/user'),
        ])
        const availableData = await api('/champions?all=false')
        setAvailableChampions(availableData.champs || [])
        setAllChampions(champData.champs || [])
        setActions(gameData.actions || [])
        if (gameData.side) {
          setSide(gameData.side)
        }

        // Start with five empty profiles
        const loadedProfiles = createEmptyProfiles()

        // Populate profiles from existing database users
        (userData||[]).forEach(user => {
          const role = user.assignedRole?.toLowerCase()

          if (!loadedProfiles[role]) return

          loadedProfiles[role] = {
            displayName: user.displayName || '',
            username: user.username || '',
            assignedRole: role,
            preferredChampions: (
              user.preferredChampions || []
            ).map(champ => champ.toLowerCase()),
          }
        })

        setProfiles(loadedProfiles)

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
  function updateProfile(role, updates) {
    setProfiles(previous => ({
      ...previous,
      [role]: {
        ...previous[role],
        ...updates,
      },
    }))
  }

  // function updatePickRole(action, role) {
  //   setActions(previous =>
  //     previous.map(item =>
  //       item === action
  //         ? { ...item, role }
  //         : item
  //     )
  //   )
  // }

  async function saveProfile(event) {
    event.preventDefault()
    setError('')

    try {
      const configuredProfiles = Object.values(profiles).filter(
        profile => profile.username.trim() !== ''
      )

      await Promise.all(
        configuredProfiles.map(profile =>
          api('/user', {
            method: 'POST',
            body: JSON.stringify(profile),
          })
        )
      )

      await api(`/side?side=${side}`, {
        method: 'POST',
      })

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
    if (!selectedSlot || pickLoading) return

    const { actionSide, type, role } = selectedSlot

    if (type === 'pick' && !role) {
      setError('Please select a role for this champion.')
      return
    }

    const action = {
      side: actionSide,
      type,
      champion: champion.toLowerCase(),
      ...(type === 'pick' && { role }),
    }

    const isOpponentPick =
      actionSide === opponentSide && type === 'pick'

    try {
      setError('')

      // Close the champion picker immediately
      setSelectedSlot(null)

      // Show loading screen for opponent picks
      if (isOpponentPick) {
        setPickLoading(true)
      }

      const result = await api('/action', {
        method: 'POST',
        body: JSON.stringify({ action }),
      })

      setActions(previous => [
        ...previous,
        {
          ...action,
          winRates: result || {},
        },
      ])

      const availableData = await api('/champions?all=false')
      setAvailableChampions(availableData.champs || [])

    } catch (err) {
      setError(err.message)

    } finally {
      setPickLoading(false)
    }
  }


  async function resetGame() {
    try {
      setError('')

      await api('/reset', {
        method: 'POST',
      })

      setActions([])
      setSelectedSlot(null)
      setAvailableChampions(allChampions)
    } catch (err) {
      setError(err.message)
    }
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
                          role: 'mid',
                        })
                      }
                    }}
                    disabled={!!action}
                  >
                    {action ? (
                      <>
                        <span className="champion-name">
                          {allChampions.find(
                            c => c.toLowerCase() === action.champion
                          ) || action.champion}
                        </span>

                        {action.type === 'pick' && (
                          <span className="champion-role">
                            {action.role?.toUpperCase()}
                          </span>
                        )}

                        <span className="slot-number">
                          {index + 1}
                        </span>

                        {action.winRates &&
                          Object.entries(action.winRates)
                            .sort((a, b) => b[1] - a[1])
                            .slice(0, 3)
                            .map(([champion, rate]) => (
                              <span key={champion} className="action-matchup">
                                {champion} – {rate.toFixed(1)}%
                              </span>
                            ))
                        }
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
      {pickLoading && (
        <div className="loading-overlay">
          <div className="loading-card">
            <div className="loading-spinner"></div>

            <h2>Analysing opponent pick</h2>

            <p>
              Calculating champion matchups and win rates...
            </p>
          </div>
        </div>
      )}
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
            <p>
              Configure a player profile for each position.
            </p>
          </div>

          <form className="settings-card" onSubmit={saveProfile}>

            {/* Role tabs */}
            <div className="role-tabs">
              {ROLES.map(role => (
                <button
                  key={role}
                  type="button"
                  className={`role-tab ${activeRole === role ? 'active' : ''
                    }`}
                  onClick={() => setActiveRole(role)}
                >
                  {role.toUpperCase()}

                  {profiles[role].username && (
                    <span className="tab-indicator">✓</span>
                  )}
                </button>
              ))}
            </div>

            {/* Active player profile */}
            {(() => {
              const profile = profiles[activeRole]

              return (
                <div className="player-profile">
                  <div className="profile-heading">
                    <div>
                      <span className="eyebrow">
                        PLAYER PROFILE
                      </span>
                      <h2>{activeRole.toUpperCase()}</h2>
                    </div>
                    <span className="profile-role">
                      {activeRole.toUpperCase()}
                    </span>
                  </div>

                  <label>
                    Display name
                    <input
                      value={profile.displayName}
                      onChange={e =>
                        updateProfile(activeRole, {
                          displayName: e.target.value,
                        })
                      }
                      placeholder="Enter display name"
                    />
                  </label>

                  <label>
                    Username
                    <input
                      value={profile.username}
                      onChange={e =>
                        updateProfile(activeRole, {
                          username: e.target.value,
                        })
                      }
                      placeholder="Enter username"
                    />
                  </label>

                  <div className="field">
                    <label>Preferred champions</label>

                    <p className="field-description">
                      Select up to five preferred champions for this role.
                    </p>

                    <div className="champion-picker-list">
                      {allChampions.map(champion => {
                        const key = champion.toLowerCase()

                        const selected =
                          profile.preferredChampions.includes(key)

                        return (
                          <button
                            type="button"
                            key={champion}
                            className={`champion-choice ${selected ? 'selected' : ''
                              }`}
                            onClick={() => {
                              updateProfile(activeRole, {
                                preferredChampions: selected
                                  ? profile.preferredChampions.filter(
                                    c => c !== key
                                  )
                                  : profile.preferredChampions.length < 5
                                    ? [...profile.preferredChampions, key]
                                    : profile.preferredChampions,
                              })
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
                </div>
              )
            })()}

            {/* Profile progress */}
            <div className="profile-progress">
              <span>
                {Object.values(profiles).filter(
                  p => p.username.trim() !== ''
                ).length} / 5 profiles configured
              </span>
            </div>

            <button className="primary-button" type="submit">
              Save all player profiles
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
              <button
                className="secondary-button"
                onClick={resetGame}
              >
                ↻ Reset Draft
              </button>
            </div>

            <div className="player-summary">
              <span className="toolbar-label">PLAYER</span>
              <strong>
                {Object.values(profiles).filter(p => p.username).length} / 5 players configured
              </strong>
              <span>PLAYER PROFILES</span>
            </div>
          </div>

          <div className="draft-board">
            <TeamPanel teamSide={side} />
            <div className="board-divider">
              <span>VS</span>
            </div>
            <TeamPanel teamSide={opponentSide} />
          </div>

        </main>
      )}

      {selectedSlot && (
        <div
          className="modal-backdrop"
          onClick={() => setSelectedSlot(null)}
        >
          <div className="champion-modal" onClick={e => e.stopPropagation()}>
            <div className="modal-heading">
              {selectedSlot.type === 'pick' && (
                <div className="role-selector">
                  <label htmlFor="pick-role">
                    CHAMPION ROLE
                  </label>

                  <div className="role-select-wrapper">
                    <span className="role-icon">⚔</span>

                    <select
                      id="pick-role"
                      value={selectedSlot.role}
                      onChange={e =>
                        setSelectedSlot(previous => ({
                          ...previous,
                          role: e.target.value,
                        }))
                      }
                    >
                      {ROLES.map(role => (
                        <option key={role} value={role}>
                          {role.charAt(0).toUpperCase() + role.slice(1)}
                        </option>
                      ))}
                    </select>

                    <span className="role-chevron">⌄</span>
                  </div>

                  <p className="role-description">
                    Choose the position this champion will play.
                  </p>
                </div>
              )}
              <div>
                <span className="eyebrow">CHAMPION SELECT</span>
                <h2>
                  Select a {selectedSlot.type}
                </h2>
              </div>
              <button onClick={() => setSelectedSlot(null)}>✕</button>
            </div>

            <div className="champion-picker-list modal-list">
              {availableChampions.map(champion => (
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
