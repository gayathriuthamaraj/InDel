
import React from "react";
import { motion } from "framer-motion";
import Navbar from "../app/components/Navbar";
import "../app/globals.css";



const DASHBOARD_STATS = [
  { label: "Active Workers", value: "2,100+" },
  { label: "Cities", value: "14" },
  { label: "Avg. Payout Time", value: "2 min" },
  { label: "Disruptions Detected", value: "1,200+" }
];

const ACCESS_LINKS = {
  worker: "https://github.com/Shravanthi20/InDel#worker-app", // Android only
  platform: "https://indel-platform-dashboard.onrender.com",
  insurer: "https://indel-urn9.onrender.com",
};

export default function App() {
  return (
    <>
      <Navbar />
      <main className="min-h-screen text-text-primary font-sans pt-16 bg-background">
        {/* 1. Hero Section (Animated) */}
        <motion.section
          className="w-full min-h-[80vh] flex flex-col justify-center items-center pb-16 pt-28 bg-primary-light"
          initial="hidden"
          animate="visible"
          variants={{
            hidden: {},
            visible: {
              transition: { staggerChildren: 0.18 }
            }
          }}
        >
          <motion.h1
            className="text-5xl md:text-6xl font-semibold leading-tight tracking-tight text-primary-dark text-center max-w-4xl"
            variants={{ hidden: { opacity: 0, y: 40 }, visible: { opacity: 1, y: 0, transition: { duration: 0.7, ease: 'easeOut' } } }}
          >
            AI-powered income protection<br />for gig workers
          </motion.h1>
          <motion.p
            className="mt-6 text-lg text-text-secondary max-w-2xl mx-auto leading-relaxed text-center"
            variants={{ hidden: { opacity: 0, y: 30 }, visible: { opacity: 1, y: 0, transition: { duration: 0.7, ease: 'easeOut', delay: 0.1 } } }}
          >
            When disruptions hit, income drops. InDel uses AI and parametric insurance to automatically protect delivery workers—no claims, no delays.
          </motion.p>

          <motion.div
            className="mt-8 flex gap-4 justify-center"
            variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0, transition: { duration: 0.6, ease: 'easeOut', delay: 0.2 } } }}
          >
            <a href="#demo" className="px-6 py-3 rounded-xl bg-primary text-surface shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200 active:scale-95">Watch Demo</a>
            <a href="#ecosystem" className="px-6 py-3 rounded-xl border border-primary text-primary-dark hover:bg-primary-light transition-all duration-200 active:scale-95">Explore Product</a>
          </motion.div>
          <div className="mt-16 flex justify-center w-full gap-6 flex-wrap">
            {[
              {
                title: "1. Register & Onboard",
                desc: "Worker signs up in the app and selects a plan."
              },
              {
                title: "2. AI Risk Scoring",
                desc: "ML model calculates fair premium based on real data."
              },
              {
                title: "3. Disruption Detection",
                desc: "Weather, AQI, or civic triggers are monitored in real time."
              },
              {
                title: "4. Automated Payout",
                desc: "If disruption hits, payout is credited instantly—no claim needed."
              }
            ].map((step, i) => (
              <motion.div
                key={i}
                className="bg-surface rounded-xl shadow p-6 flex flex-col items-center text-center border-t-4 border-primary transition-all duration-300 hover:scale-105 hover:shadow-xl"
                whileHover={{ scale: 1.05, boxShadow: '0 8px 32px rgba(0,115,157,0.12)' }}
              >
                <span className="text-3xl font-bold text-primary mb-2">{step.title}</span>
                <span className="text-text-secondary">{step.desc}</span>
              </motion.div>
            ))}
          </div>
        </motion.section>

        {/* 5. Technology Layer Section (Scroll Reveal) */}
        <motion.section
          className="max-w-5xl mx-auto py-24 px-4 bg-surface"
          id="tech"
          initial={{ opacity: 0, y: 60 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.7, ease: 'easeOut' }}
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-8 text-primary-dark">Technology Layer</h2>
          <div className="grid md:grid-cols-4 gap-6">
            {[
              { title: "AI Risk Scoring", desc: "Personalized pricing using ML models" },
              { title: "Parametric Triggers", desc: "Weather, AQI, and event APIs" },
              { title: "Fraud Detection", desc: "Multi-layer ML fraud stack" },
              { title: "Claims Automation", desc: "No manual claims—payouts are instant" },
            ].map((tech, i) => (
              <motion.div
                key={i}
                className="bg-primary-light rounded-xl shadow p-6 flex flex-col items-center text-center transition-all duration-300 hover:scale-105 hover:shadow-xl"
                whileHover={{ scale: 1.05, boxShadow: '0 8px 32px rgba(0,115,157,0.10)' }}
              >
                <span className="text-xl font-bold text-primary-dark mb-2">{tech.title}</span>
                <span className="text-text-secondary">{tech.desc}</span>
              </motion.div>
            ))}
          </div>
        </motion.section>

        {/* 6. Product Ecosystem Section (Scroll Reveal) */}
        <motion.section
          className="max-w-5xl mx-auto py-24 px-4 bg-surface"
          id="ecosystem"
          initial={{ opacity: 0, y: 60 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.7, ease: 'easeOut' }}
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-8 text-primary-dark">Product Ecosystem</h2>
          <div className="grid md:grid-cols-3 gap-6">
            {[
              {
                title: "Worker App",
                desc: "Android app for gig workers",
                link: ACCESS_LINKS.worker,
                btn: "View on GitHub"
              },
              {
                title: "Platform Dashboard",
                desc: "Operations dashboard for platforms",
                link: ACCESS_LINKS.platform,
                btn: "Open Dashboard"
              },
              {
                title: "Insurer Dashboard",
                desc: "Claims and risk dashboard for insurers",
                link: ACCESS_LINKS.insurer,
                btn: "Open Dashboard"
              }
            ].map((item, i) => (
              <motion.div
                key={i}
                className="bg-surface rounded-xl shadow p-8 flex flex-col items-center text-center border-t-4 border-primary transition-all duration-300 hover:scale-105 hover:shadow-xl"
                whileHover={{ scale: 1.05, boxShadow: '0 8px 32px rgba(0,115,157,0.10)' }}
              >
                <span className="text-xl font-bold text-primary-dark mb-2">{item.title}</span>
                <span className="text-text-secondary mb-4">{item.desc}</span>
                <a href={item.link} target="_blank" rel="noopener noreferrer" className="mt-auto px-5 py-2 rounded-lg bg-primary text-surface font-semibold shadow hover:bg-primary-dark transition">{item.btn}</a>
              </motion.div>
            ))}
          </div>
        </motion.section>

        {/* 7. Dashboard Preview Section (Scroll Reveal) */}
        <motion.section
          className="max-w-5xl mx-auto py-24 px-4 bg-surface"
          id="dashboard-preview"
          initial={{ opacity: 0, y: 60 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.7, ease: 'easeOut' }}
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-8 text-primary-dark">Dashboard Preview</h2>
          <div className="grid md:grid-cols-4 gap-6">
            {DASHBOARD_STATS.map((stat, i) => (
              <div key={i} className="bg-surface rounded-xl shadow p-6 flex flex-col items-center text-center border-t-4 border-primary">
                <span className="text-2xl font-bold text-primary-dark mb-2">{stat.label}</span>
                <span className="text-3xl text-text-primary font-mono mb-1">{stat.value}</span>
              </div>
            ))}
          </div>
        </motion.section>
        <motion.section
          className="max-w-5xl mx-auto py-24 px-4 bg-surface"
          id="demo"
          initial={{ opacity: 0, y: 60 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.7, ease: 'easeOut' }}
        >
          <h2 className="text-2xl md:text-3xl font-bold mb-8 text-primary-dark">Product Demo</h2>
          <div className="aspect-w-16 aspect-h-9 w-full rounded-xl overflow-hidden shadow-lg mb-6" style={{position: 'relative', paddingBottom: '56.25%', height: 0}}>
            <iframe
              src="https://www.youtube.com/embed/R1_1X-f7-MM"
              title="InDel Demo"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
              style={{position: 'absolute', top: 0, left: 0, width: '100%', height: '100%'}}
            />
          </div>
          <p className="text-text-secondary text-center max-w-2xl mx-auto">
            See how InDel works in action: onboarding, disruption detection, and automated payouts—all in under 2 minutes.
          </p>
        </motion.section>
      </main>
    </>
  );
}
