import type { DailyStat, DateRange, DayGroup, Meal, MealCategory, PeriodKey } from './types';
import type { Locale } from '../../../i18n/messages';
import { dictionaries } from '../../../i18n/messages';

export const MEAL_TIMEZONE = 'Europe/Kyiv';

export function getPeriodDayCount(period: PeriodKey): number {
  if (period === '7d') return 7;
  if (period === '14d') return 14;
  if (period === '30d') return 30;
  return 0;
}

export function dayKeyInKyiv(isoOrDate: string | Date): string {
  const date = typeof isoOrDate === 'string' ? new Date(isoOrDate) : isoOrDate;
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: MEAL_TIMEZONE,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date);
}

export function todayKeyInKyiv(): string {
  return dayKeyInKyiv(new Date());
}

export function shiftDayKey(dayKey: string, deltaDays: number): string {
  const [y, m, d] = dayKey.split('-').map(Number);
  const utc = new Date(Date.UTC(y, m - 1, d));
  utc.setUTCDate(utc.getUTCDate() + deltaDays);
  const yy = utc.getUTCFullYear();
  const mm = String(utc.getUTCMonth() + 1).padStart(2, '0');
  const dd = String(utc.getUTCDate()).padStart(2, '0');
  return `${yy}-${mm}-${dd}`;
}

export function dayKeyToRangeISO(fromKey: string, toKeyInclusive: string): DateRange {
  return {
    from: `${fromKey}T00:00:00+03:00`,
    to: `${shiftDayKey(toKeyInclusive, 1)}T00:00:00+03:00`,
  };
}

export function getPeriodRange(
  period: PeriodKey,
  custom?: { from: string; to: string },
): DateRange {
  if (period === 'custom' && custom?.from && custom?.to) {
    return dayKeyToRangeISO(custom.from, custom.to);
  }

  const days = getPeriodDayCount(period) || 7;
  const toKey = todayKeyInKyiv();
  const fromKey = shiftDayKey(toKey, -(days - 1));
  return dayKeyToRangeISO(fromKey, toKey);
}

export type PeriodAverages = {
  calories: number;
  protein: number;
  fat: number;
  carbs: number;
  days: number;
};

export function computePeriodAverages(summary: {
  total_calories: number;
  total_protein: number;
  total_fat: number;
  total_carbs: number;
  daily: DailyStat[];
}): PeriodAverages {
  const days = summary.daily.filter((d) => d.meals > 0).length || summary.daily.length;
  if (days <= 0) {
    return { days: 0, calories: 0, protein: 0, fat: 0, carbs: 0 };
  }
  return {
    days,
    calories: summary.total_calories / days,
    protein: summary.total_protein / days,
    fat: summary.total_fat / days,
    carbs: summary.total_carbs / days,
  };
}

export function fillDailySeries(
  fromKey: string,
  toKeyInclusive: string,
  daily: DailyStat[],
): DailyStat[] {
  const byDay = new Map<string, DailyStat>();
  for (const item of daily) {
    byDay.set(dayKeyInKyiv(item.date), item);
  }

  const out: DailyStat[] = [];
  let cursor = fromKey;
  while (cursor <= toKeyInclusive) {
    const existing = byDay.get(cursor);
    out.push(
      existing ?? {
        date: `${cursor}T00:00:00+03:00`,
        calories: 0,
        protein: 0,
        fat: 0,
        carbs: 0,
        meals: 0,
      },
    );
    cursor = shiftDayKey(cursor, 1);
  }
  return out;
}

export function formatNumber(value: number, digits = 0, locale: Locale = 'uk'): string {
  const tag = locale === 'en' ? 'en-US' : 'uk-UA';
  return new Intl.NumberFormat(tag, {
    maximumFractionDigits: digits,
    minimumFractionDigits: digits,
  }).format(value);
}

export function formatDayLabel(isoDate: string, locale: Locale = 'uk'): string {
  const date = new Date(isoDate);
  const tag = locale === 'en' ? 'en-US' : 'uk-UA';
  return new Intl.DateTimeFormat(tag, {
    timeZone: MEAL_TIMEZONE,
    weekday: 'short',
    day: '2-digit',
    month: 'short',
  }).format(date);
}

export function formatDateTime(isoDate: string, locale: Locale = 'uk'): string {
  const date = new Date(isoDate);
  const tag = locale === 'en' ? 'en-US' : 'uk-UA';
  return new Intl.DateTimeFormat(tag, {
    timeZone: MEAL_TIMEZONE,
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

export function categoryLabel(category: MealCategory | string, locale: Locale = 'uk'): string {
  const dict = dictionaries[locale];
  switch (category) {
    case 'high_quality_protein':
      return dict.catProtein;
    case 'long_acting_carbs':
      return dict.catSlowCarbs;
    case 'lipids':
      return dict.catLipids;
    case 'fast_acting_carbs':
      return dict.catFastCarbs;
    default:
      return category;
  }
}

export function dayKeyFromISO(isoDate: string): string {
  return dayKeyInKyiv(isoDate);
}

export function groupMealsByDay(meals: Meal[], locale: Locale = 'uk'): DayGroup[] {
  const map = new Map<string, Meal[]>();
  for (const meal of meals) {
    const key = dayKeyFromISO(meal.created_at);
    const list = map.get(key) ?? [];
    list.push(meal);
    map.set(key, list);
  }

  return Array.from(map.entries())
    .sort((a, b) => (a[0] < b[0] ? 1 : -1))
    .map(([dateKey, dayMeals]) => ({
      dateKey,
      label: formatDayLabel(dayMeals[0].created_at, locale),
      meals: [...dayMeals].sort((a, b) => a.sort_order - b.sort_order || a.id - b.id),
    }));
}

export function toDatetimeLocalValue(isoDate: string): string {
  const date = new Date(isoDate);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function fromDatetimeLocalValue(value: string): string {
  return new Date(value).toISOString();
}
