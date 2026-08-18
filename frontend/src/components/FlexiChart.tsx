import { useMemo } from "react";
import {
  Area,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  ResponsiveContainer,
  XAxis,
  YAxis,
} from "recharts";
import type { Entry, Term } from "../lib/api";
import { calculateStats } from "../lib/stats";

interface FlexiChartProps {
  entries: Entry[];
  term: Term;
}

type ChartPoint = {
  timestamp: number;
  actualAmount: number | null;
  projectedAmount: number | null;
};

const GREEN = "#4ade80";
const GREEN_FILL = "rgba(74, 222, 128, 0.12)";
const GRID = "#404040";
const LABEL = "#ffffff80";
const PROJECTED = "rgba(255, 255, 255, 0.45)";

/**
 * Converts a YYYY-MM-DD date string to milliseconds since epoch.
 *
 * Recharts' time axis expects numeric timestamps, but our dates are stored
 * as strings.
 *
 * Parsing the components manually keeps the date in local time;
 * `new Date("YYYY-MM-DD")` would treat it as UTC and can shift the day.
 */
function dateToMs(dateStr: string): number {
  const [year, month, day] = dateStr.split("-").map(Number);
  return new Date(year, month - 1, day).getTime();
}

/**
 * Given the start and end date strings, returns an array of all the ticks that
 * should be visible on the x-axis.
 *
 * Includes the start date and end date. Plus every 1st and 15th day of the
 * month in between.
 */
function xAxisTicks(start: string, end: string): number[] {
  const minMs = dateToMs(start);
  const maxMs = dateToMs(end);
  const ticks = new Set<number>([minMs, maxMs]);
  const current = new Date(minMs);
  current.setDate(1);

  while (current.getTime() <= maxMs) {
    for (const day of [1, 15]) {
      const t = new Date(
        current.getFullYear(),
        current.getMonth(),
        day,
      ).getTime();
      if (t >= minMs && t <= maxMs) ticks.add(t);
    }
    current.setMonth(current.getMonth() + 1);
  }

  return [...ticks].sort((a, b) => a - b);
}

/**
 * Adds the given number of days to a YYYY-MM-DD date string.
 *
 * Returns a new date string in YYYY-MM-DD format (local time).
 */
function addDays(dateStr: string, days: number): string {
  const [year, month, day] = dateStr.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  date.setDate(date.getDate() + Math.ceil(days));
  return date.toLocaleDateString("en-CA");
}

function buildChartPoints(entries: Entry[], term: Term): ChartPoint[] {
  const stats = calculateStats(entries, term);

  // One amount per date: term start, then each entry
  const amountByDate = new Map<string, number>([
    [term.startDate, term.startingAmount],
  ]);
  for (const entry of entries) {
    amountByDate.set(entry.date, entry.amountRemaining);
  }

  const dates = [...amountByDate.keys()].sort();
  const lastDate = dates[dates.length - 1];
  const lastAmount = amountByDate.get(lastDate)!;

  // Dashed projection line goes from last entry to
  // either term end or run out date
  let projectionDate = term.endDate;
  let projectionAmount = Math.max(stats.endingProjection, 0);
  if (stats.endingProjection <= 0 && stats.dailySpend > 0) {
    // TODO: if we prevent a term from being able to start with $0,
    // then we don't need to check for dailySpend being non-zero
    projectionDate = addDays(lastDate, lastAmount / stats.dailySpend);
  }
  projectionDate =
    projectionDate > term.endDate ? term.endDate : projectionDate;

  const pointDates = [...new Set([...dates, projectionDate])].sort();

  // null means "don't draw here" for that series.
  return pointDates.map((date) => ({
    timestamp: dateToMs(date),
    actualAmount: amountByDate.get(date) ?? null,
    projectedAmount:
      date === lastDate
        ? lastAmount
        : date === projectionDate
          ? projectionAmount
          : null,
  }));
}

/**
 * Determines out what the highest number on the y-axis should be.
 *
 * Based off the max amount in the given points.
 */
function chartYMax(points: ChartPoint[]): number {
  // ChartPoint has both actualAmount and projectedAmount, where one is null
  // Filter out nulls
  const values = points.flatMap((point) =>
    [point.actualAmount, point.projectedAmount].filter(
      (value): value is number => value != null,
    ),
  );
  const max = Math.max(...values, 0);
  return Math.ceil((max * 1.25) / 200) * 200;
}

export default function FlexiChart({
  entries,
  term,
}: Readonly<FlexiChartProps>) {
  const points = useMemo(
    () => buildChartPoints(entries, term),
    [entries, term],
  );
  const yMax = useMemo(() => chartYMax(points), [points]);
  const xTicks = useMemo(
    () => xAxisTicks(term.startDate, term.endDate),
    [term.endDate, term.startDate],
  );

  return (
    <div className="rounded-lg border border-(--border) bg-(--background-light) p-4">
      <h2 className="mb-4 text-base font-semibold">
        Flexi Remaining Over Time
      </h2>
      <div className="h-[280px] w-full">
        <ResponsiveContainer width="100%" height="100%">
          <ComposedChart
            data={points}
            margin={{ top: 8, right: 8, bottom: 0, left: 0 }}
          >
            <CartesianGrid vertical={false} stroke={GRID} />
            <XAxis
              dataKey="timestamp"
              type="number"
              scale="time"
              domain={[dateToMs(term.startDate), dateToMs(term.endDate)]}
              ticks={xTicks}
              tick={{ fill: LABEL, fontSize: 12 }}
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => {
                const date = new Date(value);
                return `${date.getMonth() + 1}/${date.getDate()}`;
              }}
            />
            <YAxis
              domain={[0, yMax]}
              tick={{ fill: LABEL, fontSize: 12 }}
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => `$${value}`}
              width={52}
            />
            <Legend
              verticalAlign="bottom"
              iconType="plainline"
              wrapperStyle={{ fontSize: 12, color: LABEL, paddingTop: 8 }}
            />
            <Area
              type="monotone"
              dataKey="actualAmount"
              name="Actual Remaining"
              stroke={GREEN}
              strokeWidth={2}
              fill={GREEN_FILL}
              connectNulls
              dot={false}
              activeDot={false}
              isAnimationActive={false}
            />
            <Line
              type="monotone"
              dataKey="actualAmount"
              legendType="none"
              stroke="none"
              connectNulls
              dot={(props) => {
                const { cx, cy, value } = props;
                if (cx == null || cy == null || value == null) return null;
                return <circle cx={cx} cy={cy} r={5} fill={GREEN} />;
              }}
              activeDot={false}
              isAnimationActive={false}
            />
            <Line
              type="monotone"
              dataKey="projectedAmount"
              name="Projected Remaining"
              stroke={PROJECTED}
              strokeWidth={2}
              strokeDasharray="6 6"
              connectNulls
              dot={false}
              activeDot={false}
              isAnimationActive={false}
            />
          </ComposedChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
