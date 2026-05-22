import { describe, it, expect } from 'vitest';
import { StepCounter } from './stepCounter';

describe('StepCounter', () => {
  it('ignores below-threshold magnitudes', () => {
    const c = new StepCounter({ threshold: 1.2 });
    c.feed(0.5, 0.5, 0.5, 0);
    expect(c.count).toBe(0);
  });

  it('counts peaks separated by windowMs', () => {
    const c = new StepCounter({ threshold: 1.2, windowMs: 250 });
    c.feed(1, 1, 1, 0); // mag = sqrt(3) ≈ 1.732
    c.feed(1, 1, 1, 100); // within window — skip
    c.feed(1, 1, 1, 260); // outside window — count
    c.feed(1, 1, 1, 520);
    expect(c.count).toBe(3);
  });

  it('reset zeroes counter', () => {
    const c = new StepCounter({ threshold: 1.2 });
    c.feed(1, 1, 1, 0);
    c.reset();
    expect(c.count).toBe(0);
  });

  it('emits onStep callback', () => {
    const events: number[] = [];
    const c = new StepCounter({ threshold: 1, onStep: (e) => events.push(e.t) });
    c.feed(1, 1, 1, 100);
    c.feed(1, 1, 1, 600);
    expect(events).toEqual([100, 600]);
  });
});
