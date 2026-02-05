/**
 * Global Gesture Detection Utility
 * Provides edge swipe detection for mobile devices and event communication
 * between components that need to respond to gestures.
 */

interface GestureEvent {
  type: "left-edge-swipe-right" | "right-edge-swipe-left" | "other";
  startX: number;
  startY: number;
  endX: number;
  endY: number;
  deltaX: number;
  deltaY: number;
  duration: number;
  velocity: number;
}

type GestureCallback = (event: GestureEvent) => void;

class GestureDetector {
  private subscribers: Map<string, GestureCallback[]> = new Map();
  private isMobile: boolean = false;
  private isEnabled: boolean = true;

  // Touch tracking
  private touchStartX: number = 0;
  private touchStartY: number = 0;
  private touchStartTime: number = 0;
  private isSwiping: boolean = false;

  // Configuration
  private edgeThreshold: number = 30; // pixels from edge
  private swipeThreshold: number = 50; // minimum swipe distance
  private maxSwipeTime: number = 500; // maximum swipe duration (ms)
  private verticalThreshold: number = 20; // maximum vertical deviation

  constructor() {
    console.log("[GestureDetector] Initializing global gesture detection");
    console.log(
      "[GestureDetector] Screen width:",
      window.innerWidth,
      "isMobile:",
      window.innerWidth <= 736
    );
    this.isMobile = window.innerWidth <= 736;
    this.setupEventListeners();
    console.log("[GestureDetector] Gesture detection initialized successfully");
  }

  /**
   * Subscribe to gesture events
   * @param eventType Type of gesture to listen for
   * @param callback Function to call when gesture occurs
   * @returns Unsubscribe function
   */
  subscribe(eventType: string, callback: GestureCallback): () => void {
    console.log(`[GestureDetector] Subscribing to ${eventType}`);

    if (!this.subscribers.has(eventType)) {
      this.subscribers.set(eventType, []);
    }

    const callbacks = this.subscribers.get(eventType)!;
    callbacks.push(callback);

    // Return unsubscribe function
    return () => {
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
        console.log(`[GestureDetector] Unsubscribed from ${eventType}`);
      }
    };
  }

  /**
   * Emit gesture event to all subscribers
   * @param event Gesture event data
   */
  private emit(event: GestureEvent): void {
    console.log(`[GestureDetector] Emitting ${event.type} event`, {
      deltaX: event.deltaX,
      deltaY: event.deltaY,
      duration: event.duration,
      velocity: event.velocity,
    });

    const callbacks = this.subscribers.get(event.type);
    if (callbacks) {
      callbacks.forEach((callback) => {
        try {
          callback(event);
        } catch (error) {
          console.error(`[GestureDetector] Error in gesture callback:`, error);
        }
      });
    }
  }

  /**
   * Setup global touch event listeners
   */
  private setupEventListeners(): void {
    if (typeof window === "undefined") return;

    console.log("[GestureDetector] Setting up global event listeners");

    document.addEventListener("touchstart", this.handleTouchStart.bind(this), {
      passive: true,
    });
    document.addEventListener("touchmove", this.handleTouchMove.bind(this), {
      passive: false,
    });
    document.addEventListener("touchend", this.handleTouchEnd.bind(this), {
      passive: true,
    });

    // Update mobile detection on resize
    window.addEventListener("resize", () => {
      const wasMobile = this.isMobile;
      this.isMobile = window.innerWidth <= 736;
      if (wasMobile !== this.isMobile) {
        console.log(
          `[GestureDetector] Mobile detection changed: ${wasMobile} -> ${this.isMobile}`
        );
      }
    });
  }

  /**
   * Handle touch start events
   */
  private handleTouchStart(event: TouchEvent): void {
    if (!this.isMobile || !this.isEnabled) {
      console.log(
        `[GestureDetector] Ignoring touch: mobile=${this.isMobile}, enabled=${this.isEnabled}`
      );
      return;
    }

    if (event.touches.length !== 1) {
      console.log(
        `[GestureDetector] Ignoring multi-touch: ${event.touches.length} touches`
      );
      return;
    }

    const touch = event.touches[0];
    this.touchStartX = touch.clientX;
    this.touchStartY = touch.clientY;
    this.touchStartTime = Date.now();
    this.isSwiping = false;

    console.log(
      `[GestureDetector] Touch started at (${this.touchStartX}, ${this.touchStartY})`,
      {
        time: this.touchStartTime,
        screenWidth: window.innerWidth,
      }
    );
  }

  /**
   * Handle touch move events
   */
  private handleTouchMove(event: TouchEvent): void {
    if (!this.isMobile || !this.isEnabled || !this.touchStartX) {
      console.log(
        `[GestureDetector] Ignoring touch move: mobile=${this.isMobile}, enabled=${this.isEnabled}, hasStart=${!!this.touchStartX}`
      );
      return;
    }

    if (event.touches.length !== 1) {
      console.log(
        `[GestureDetector] Ignoring multi-touch move: ${event.touches.length} touches`
      );
      return;
    }

    // Check if touch started near left edge
    if (this.touchStartX > this.edgeThreshold) {
      console.log(
        `[GestureDetector] Touch started too far from edge (${this.touchStartX} > ${this.edgeThreshold})`
      );
      return;
    }

    const touch = event.touches[0];
    const deltaX = touch.clientX - this.touchStartX;
    const deltaY = Math.abs(touch.clientY - this.touchStartY);

    console.log(
      `[GestureDetector] Touch move: deltaX=${deltaX}, deltaY=${deltaY}, currentX=${touch.clientX}`
    );

    // Prevent vertical swipes
    if (deltaY > this.verticalThreshold) {
      console.log(
        `[GestureDetector] Vertical movement too large (${deltaY} > ${this.verticalThreshold}) - rejecting`
      );
      return;
    }

    // Only handle rightward swipes
    if (deltaX < 0) {
      console.log(
        `[GestureDetector] Leftward swipe detected (${deltaX} < 0) - rejecting`
      );
      return;
    }

    // Swipe distance is significant enough
    if (deltaX > 10) {
      event.preventDefault();
      this.isSwiping = true;
      console.log(
        `[GestureDetector] Valid swiping detected: deltaX=${deltaX}, preventing default`
      );
    }
  }

  /**
   * Handle touch end events
   */
  private handleTouchEnd(event: TouchEvent): void {
    if (!this.isMobile || !this.isEnabled || !this.isSwiping) {
      console.log(
        `[GestureDetector] Ignoring touch end: mobile=${this.isMobile}, enabled=${this.isEnabled}, swiping=${this.isSwiping}`
      );
      return;
    }

    const touch = event.changedTouches[0];
    const deltaX = touch.clientX - this.touchStartX;
    const deltaY = touch.clientY - this.touchStartY;
    const duration = Date.now() - this.touchStartTime;
    const velocity = deltaX / duration;

    console.log(
      `[GestureDetector] Touch ended: deltaX=${deltaX}, duration=${duration}ms, velocity=${velocity.toFixed(2)}px/ms`
    );

    // Check if swipe meets criteria
    if (deltaX > this.swipeThreshold && duration < this.maxSwipeTime) {
      const gestureEvent: GestureEvent = {
        type: "left-edge-swipe-right",
        startX: this.touchStartX,
        startY: this.touchStartY,
        endX: touch.clientX,
        endY: touch.clientY,
        deltaX,
        deltaY,
        duration,
        velocity,
      };

      this.emit(gestureEvent);
      console.log(
        `[GestureDetector] Valid left-edge-swipe-right gesture detected`
      );
    } else {
      console.log(
        `[GestureDetector] Gesture rejected: deltaX=${deltaX} <= ${this.swipeThreshold} or duration=${duration} >= ${this.maxSwipeTime}`
      );
    }

    // Reset state
    this.isSwiping = false;
  }

  /**
   * Enable or disable gesture detection
   */
  setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;
    console.log(
      `[GestureDetector] Gesture detection ${enabled ? "enabled" : "disabled"}`
    );
  }

  /**
   * Update configuration
   */
  configure(
    config: Partial<{
      edgeThreshold: number;
      swipeThreshold: number;
      maxSwipeTime: number;
      verticalThreshold: number;
    }>
  ): void {
    Object.assign(this, config);
    console.log("[GestureDetector] Configuration updated:", config);
  }

  /**
   * Get current state and statistics
   */
  getStats(): {
    isMobile: boolean;
    isEnabled: boolean;
    subscriberCount: number;
    config: Record<string, number>;
  } {
    const subscriberCount = Array.from(this.subscribers.values()).reduce(
      (total, callbacks) => total + callbacks.length,
      0
    );

    return {
      isMobile: this.isMobile,
      isEnabled: this.isEnabled,
      subscriberCount,
      config: {
        edgeThreshold: this.edgeThreshold,
        swipeThreshold: this.swipeThreshold,
        maxSwipeTime: this.maxSwipeTime,
        verticalThreshold: this.verticalThreshold,
      },
    };
  }

  /**
   * Log current state for debugging
   */
  logState(): void {
    const stats = this.getStats();
    console.log("[GestureDetector] Current state:", stats);
  }

  /**
   * Clean up event listeners
   */
  destroy(): void {
    console.log("[GestureDetector] Cleaning up event listeners");
    document.removeEventListener(
      "touchstart",
      this.handleTouchStart.bind(this)
    );
    document.removeEventListener("touchmove", this.handleTouchMove.bind(this));
    document.removeEventListener("touchend", this.handleTouchEnd.bind(this));
  }
}

// Create singleton instance
const gestureDetector = new GestureDetector();

// Export singleton and types
export default gestureDetector;
export type { GestureEvent, GestureCallback };
