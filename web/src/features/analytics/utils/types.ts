export type DailyStat = {
  date: string;
  calories: number;
  protein: number;
  fat: number;
  carbs: number;
  meals: number;
};

export type StatsSummary = {
  from: string;
  to: string;
  total_calories: number;
  total_protein: number;
  total_fat: number;
  total_carbs: number;
  meal_count: number;
  daily: DailyStat[];
};

export type MealCategory =
  'high_quality_protein' | 'long_acting_carbs' | 'lipids' | 'fast_acting_carbs';

export const MEAL_CATEGORIES: { value: MealCategory; label: string }[] = [
  { value: 'high_quality_protein', label: 'Якісний білок' },
  { value: 'long_acting_carbs', label: 'Повільні вуглеводи' },
  { value: 'lipids', label: 'Жири' },
  { value: 'fast_acting_carbs', label: 'Швидкі вуглеводи' },
];

export type MealItem = {
  id: number;
  name: string;
  weight_g: number;
  calories: number;
  protein: number;
  fat: number;
  carbs: number;
  category: MealCategory;
};

export type Meal = {
  id: number;
  total_calories: number;
  total_protein: number;
  total_fat: number;
  total_carbs: number;
  raw_text: string;
  sort_order: number;
  created_at: string;
  items: MealItem[];
};

export type MealsResponse = {
  meals: Meal[];
};

export type MealItemInput = {
  name: string;
  weight_g: number;
  calories: number;
  protein: number;
  fat: number;
  carbs: number;
  category: MealCategory;
};

export type CreateMealPayload = {
  raw_text?: string;
  created_at?: string;
  items?: MealItemInput[];
};

export type UpdateMealPayload = {
  raw_text?: string;
  created_at?: string;
  items?: MealItemInput[];
};

export type DayGroup = {
  dateKey: string;
  label: string;
  meals: Meal[];
};

export type PeriodKey = '7d' | '14d' | '30d' | 'custom';

export type DateRange = {
  from: string;
  to: string;
};
