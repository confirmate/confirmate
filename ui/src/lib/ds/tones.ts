/** The semantic tones of the design system. They map to token based classes, never to raw colors. */
export type Tone = 'neutral' | 'brand' | 'ok' | 'warn' | 'bad';

/** Soft background with readable text, for chips and tags. */
export const softTone: Record<Tone, string> = {
	neutral: 'bg-neutral-soft text-ink-soft',
	brand: 'bg-brand-soft text-brand-ink',
	ok: 'bg-ok-soft text-ok-ink',
	warn: 'bg-warn-soft text-warn-ink',
	bad: 'bg-bad-soft text-bad-ink'
};

/** Solid color, for bars and accents. */
export const solidTone: Record<Tone, string> = {
	neutral: 'bg-line-strong',
	brand: 'bg-brand',
	ok: 'bg-ok',
	warn: 'bg-warn',
	bad: 'bg-bad'
};

/** Border color, for accents of cards. */
export const borderTone: Record<Tone, string> = {
	neutral: 'border-l-line-strong',
	brand: 'border-l-brand',
	ok: 'border-l-ok',
	warn: 'border-l-warn',
	bad: 'border-l-bad'
};
