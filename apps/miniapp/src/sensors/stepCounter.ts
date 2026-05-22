// Peak-detection step counter dùng accelerometer.
// MVP: tự cài, không phụ thuộc zmp-sdk để dễ unit test trong jsdom.
// Lên prod: thay listener bằng `zmp-sdk` `onAccelerometerChange`.
//
// Thuật toán: tính độ lớn vector gia tốc, detect peak khi mag vượt threshold
// và đã đủ thời gian từ peak trước (debounce window).

export interface StepEvent {
  t: number; // epoch ms
  magnitude: number;
}

export interface StepCounterOpts {
  threshold?: number;       // ngưỡng g; default 1.2
  windowMs?: number;        // debounce; default 250
  onStep?: (e: StepEvent) => void;
}

export class StepCounter {
  // null = chưa có peak nào → peak đầu tiên luôn được tính.
  private lastPeakT: number | null = null;
  private readonly threshold: number;
  private readonly windowMs: number;
  private readonly onStep?: (e: StepEvent) => void;
  public count = 0;

  constructor(opts: StepCounterOpts = {}) {
    this.threshold = opts.threshold ?? 1.2;
    this.windowMs = opts.windowMs ?? 250;
    this.onStep = opts.onStep;
  }

  feed(x: number, y: number, z: number, tMs: number): boolean {
    const mag = Math.sqrt(x * x + y * y + z * z);
    if (mag < this.threshold) return false;
    if (this.lastPeakT !== null && tMs - this.lastPeakT < this.windowMs) return false;
    this.lastPeakT = tMs;
    this.count += 1;
    this.onStep?.({ t: tMs, magnitude: mag });
    return true;
  }

  reset() {
    this.count = 0;
    this.lastPeakT = null;
  }
}
