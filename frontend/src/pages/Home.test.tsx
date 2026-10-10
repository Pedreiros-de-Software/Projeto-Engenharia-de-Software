import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import Home from './Home'

afterEach(cleanup)

describe('Home', () => {
  it('identifica o produto para quem acessa a página inicial', () => {
    render(<Home />)
    expect(screen.getByRole('heading', { level: 1, name: 'Consensu' })).toBeVisible()
  })

  it('informa os três contextos de decisão atendidos pelo produto', () => {
    render(<Home />)
    const description = screen.getByRole('heading', { level: 3 })
    expect(description).toHaveTextContent('Eleições de CA')
    expect(description).toHaveTextContent('Condomínio')
    expect(description).toHaveTextContent('Priorização de projetos')
  })
})
