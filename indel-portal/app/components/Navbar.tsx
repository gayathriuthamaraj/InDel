import React, { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";

function scrollTo(id: string) {
  const el = document.getElementById(id);
  if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
}

const navLinks = [
  { label: "Problem", id: "problem" },
  { label: "How It Works", id: "how-it-works" },
  { label: "Features", id: "features" },
  { label: "Technology", id: "tech" },
  { label: "Demo", id: "demo" },
];

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  const [scrollProgress, setScrollProgress] = useState(0);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [activeSection, setActiveSection] = useState("");

  useEffect(() => {
    const handleScroll = () => {
      const scrollTop = window.scrollY;
      const docHeight = document.documentElement.scrollHeight - window.innerHeight;
      setScrolled(scrollTop > 40);
      setScrollProgress(docHeight > 0 ? (scrollTop / docHeight) * 100 : 0);

      // Determine active section
      const sections = navLinks.map((l) => l.id);
      for (let i = sections.length - 1; i >= 0; i--) {
        const el = document.getElementById(sections[i]);
        if (el && el.getBoundingClientRect().top <= 140) {
          setActiveSection(sections[i]);
          break;
        }
      }
    };
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <>
      {/* Scroll progress bar */}
      <div className="fixed top-0 left-0 z-[60] h-0.5 bg-primary transition-all duration-150" style={{ width: `${scrollProgress}%` }} />

      <motion.nav
        className={`fixed top-0.5 left-0 right-0 z-50 transition-all duration-300 ${
          scrolled
            ? "glass shadow-lg shadow-primary/5"
            : "bg-transparent"
        }`}
        initial={{ y: -64 }}
        animate={{ y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
      >
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          {/* Logo */}
          <button
            onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
            className={`font-bold text-xl tracking-tight transition-colors duration-300 ${scrolled ? "text-primary-dark" : "text-white"}`}
          >
            <span className="text-primary">In</span>
            <span>Del</span>
          </button>

          {/* Desktop nav */}
          <div className="hidden md:flex items-center gap-1">
            {navLinks.map((link) => (
              <button
                key={link.id}
                onClick={() => scrollTo(link.id)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                  activeSection === link.id
                    ? "bg-primary text-white"
                    : scrolled
                    ? "text-text-primary hover:bg-primary-light hover:text-primary"
                    : "text-white/80 hover:text-white hover:bg-white/10"
                }`}
              >
                {link.label}
              </button>
            ))}
          </div>

          {/* Desktop CTA */}
          <div className="hidden md:flex items-center gap-3">
            <motion.a
              href="#demo"
              onClick={(e) => { e.preventDefault(); scrollTo("demo"); }}
              className="px-5 py-2 rounded-lg bg-primary text-white text-sm font-semibold shadow-md shadow-primary/30 hover:bg-primary-dark transition-all duration-200"
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.96 }}
            >
              Watch Demo
            </motion.a>
          </div>

          {/* Mobile hamburger */}
          <button
            className={`md:hidden p-2 rounded-lg transition-colors duration-200 ${scrolled ? "text-primary-dark hover:bg-primary-light" : "text-white hover:bg-white/10"}`}
            onClick={() => setMobileOpen(!mobileOpen)}
            aria-label="Toggle menu"
          >
            <div className="w-5 flex flex-col gap-1.5">
              <span className={`block h-0.5 transition-all duration-300 ${scrolled ? "bg-primary-dark" : "bg-white"} ${mobileOpen ? "rotate-45 translate-y-2" : ""}`} />
              <span className={`block h-0.5 transition-all duration-300 ${scrolled ? "bg-primary-dark" : "bg-white"} ${mobileOpen ? "opacity-0" : ""}`} />
              <span className={`block h-0.5 transition-all duration-300 ${scrolled ? "bg-primary-dark" : "bg-white"} ${mobileOpen ? "-rotate-45 -translate-y-2" : ""}`} />
            </div>
          </button>
        </div>

        {/* Mobile menu */}
        <AnimatePresence>
          {mobileOpen && (
            <motion.div
              className="md:hidden glass border-t border-primary/10 px-6 py-4 flex flex-col gap-2"
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              exit={{ opacity: 0, height: 0 }}
              transition={{ duration: 0.25 }}
            >
              {navLinks.map((link) => (
                <button
                  key={link.id}
                  onClick={() => { scrollTo(link.id); setMobileOpen(false); }}
                  className="text-left px-4 py-3 rounded-lg text-sm font-medium text-text-primary hover:bg-primary-light hover:text-primary transition-all"
                >
                  {link.label}
                </button>
              ))}
              <div className="mt-2 pt-2 border-t border-primary/10 flex flex-col gap-2">
                {/* Platform Dashboard link removed as requested */}
                <a
                  href={import.meta.env.VITE_INSURER_DASHBOARD_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="px-4 py-3 rounded-lg text-sm font-medium bg-primary text-white text-center"
                >
                  Insurer Dashboard
                </a>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.nav>
    </>
  );
}
