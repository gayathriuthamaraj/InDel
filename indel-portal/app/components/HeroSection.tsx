import React, { useEffect, useRef, useState } from "react";
import { motion } from "framer-motion";

const stats = [
  { value: "15M+", label: "Gig Workers in India" },
  { value: "₹527", label: "Avg. Payout / Event" },
  { value: "2 min", label: "Payout Time" },
  { value: "Zero", label: "Claims to File" },
];

function AnimatedStat({ value, label, delay }: { value: string; label: string; delay: number }) {
  return (
    <motion.div
      className="flex flex-col items-center"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: delay + 0.8, duration: 0.6, ease: "easeOut" }}
    >
      <span className="text-2xl md:text-3xl font-bold text-white">{value}</span>
      <span className="text-xs md:text-sm text-white/50 mt-1 text-center">{label}</span>
    </motion.div>
  );
}

export default function HeroSection() {
  const [typedLine, setTypedLine] = useState("");
  const headline = "The rain starts at 11:40 AM.";
  const ref = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    let i = 0;
    const tick = () => {
      setTypedLine(headline.slice(0, i + 1));
      i++;
      if (i < headline.length) ref.current = setTimeout(tick, 48);
    };
    const start = setTimeout(tick, 600);
    return () => { clearTimeout(start); if (ref.current) clearTimeout(ref.current); };
  }, []);

  return (
    <section
      id="hero"
      className="relative min-h-screen flex flex-col justify-center items-center overflow-hidden dark-section"
    >
      {/* Grid overlay */}
      <div className="absolute inset-0 grid-overlay opacity-40 pointer-events-none" />

      {/* Animated blobs */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 rounded-full bg-primary/20 blur-3xl animate-blob pointer-events-none" />
      <div className="absolute top-1/3 right-1/4 w-80 h-80 rounded-full bg-blue-500/10 blur-3xl animate-blob animation-delay-2000 pointer-events-none" />
      <div className="absolute bottom-1/4 left-1/3 w-64 h-64 rounded-full bg-cyan-400/10 blur-3xl animate-blob animation-delay-4000 pointer-events-none" />

      <div className="relative z-10 max-w-5xl mx-auto px-6 text-center pt-24 pb-16">
        {/* Hackathon badge */}
        <motion.div
          initial={{ opacity: 0, y: -16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="inline-flex items-center gap-2 mb-8 px-4 py-2 rounded-full border border-primary/40 bg-primary/10 text-primary-light text-xs font-semibold tracking-widest uppercase"
        >
          <span className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse-slow" />
          Guidewire DEVTrails 2026 · Phase 3: Scale & Optimize
        </motion.div>

        {/* Typewriter line */}
        <motion.div
          className="font-mono text-lg md:text-xl text-white/60 mb-4 h-7"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.3, delay: 0.4 }}
        >
          {typedLine}
          <span className="inline-block w-0.5 h-5 bg-primary ml-1 animate-pulse" />
        </motion.div>

        {/* Main headline */}
        <motion.h1
          className="text-5xl md:text-7xl font-extrabold leading-[1.08] tracking-tight text-white mb-6"
          initial={{ opacity: 0, y: 32 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.5, ease: [0.25, 0.46, 0.45, 0.94] }}
        >
          Income protection
          <br />
          <span className="text-gradient-light">that just arrives.</span>
        </motion.h1>

        {/* Sub-headline */}
        <motion.p
          className="text-lg md:text-xl text-white/60 max-w-2xl mx-auto leading-relaxed mb-10"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, delay: 0.7, ease: "easeOut" }}
        >
          InDel is a B2B parametric insurance platform. When disruptions hit, workers
          receive verified payouts via UPI — automatically. No claims. No forms. No waiting.
        </motion.p>

        {/* CTAs */}
        <motion.div
          className="flex flex-col sm:flex-row gap-4 justify-center items-center"
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.9, ease: "easeOut" }}
        >
          <motion.a
            href="#demo"
            onClick={(e) => { e.preventDefault(); document.getElementById("demo")?.scrollIntoView({ behavior: "smooth" }); }}
            className="flex items-center gap-2 px-8 py-3.5 rounded-xl bg-primary text-white font-semibold text-sm shadow-xl shadow-primary/30 hover:bg-primary-dark transition-all"
            whileHover={{ scale: 1.04, y: -2 }}
            whileTap={{ scale: 0.97 }}
          >
            <span>▶</span> Watch Demo
          </motion.a>
          <motion.a
            href="#how-it-works"
            onClick={(e) => { e.preventDefault(); document.getElementById("how-it-works")?.scrollIntoView({ behavior: "smooth" }); }}
            className="px-8 py-3.5 rounded-xl border border-white/20 text-white/80 hover:border-white/50 hover:text-white text-sm font-semibold transition-all"
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
          >
            See How It Works →
          </motion.a>
        </motion.div>
      </div>

      {/* Stats strip */}
      <motion.div
        className="relative z-10 w-full max-w-3xl mx-auto px-6 pb-16"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.1, duration: 0.6 }}
      >
        <div className="glass-dark rounded-2xl px-8 py-6 grid grid-cols-2 md:grid-cols-4 gap-6">
          {stats.map((s, i) => (
            <AnimatedStat key={i} value={s.value} label={s.label} delay={i * 0.1} />
          ))}
        </div>
      </motion.div>

      {/* Scroll indicator */}
      <motion.div
        className="absolute bottom-6 left-1/2 -translate-x-1/2 flex flex-col items-center gap-1 text-white/30 text-xs"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.6, duration: 0.5 }}
      >
        <div className="w-5 h-8 border border-white/20 rounded-full flex justify-center pt-1.5">
          <motion.div
            className="w-1 h-1.5 bg-white/40 rounded-full"
            animate={{ y: [0, 10, 0] }}
            transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
          />
        </div>
        scroll
      </motion.div>
    </section>
  );
}
