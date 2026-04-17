import React from "react";
import { motion } from "framer-motion";
import Navbar from "../app/components/Navbar";
import HeroSection from "../app/components/HeroSection";
import ProblemSection from "../app/components/ProblemSection";
import InsightSection from "../app/components/InsightSection";
import HowItWorksSection from "../app/components/HowItWorksSection";
import FeatureSection from "../app/components/FeatureSection";
import PricingIntelSection from "../app/components/PricingIntelSection";
import PilotNumbersSection from "../app/components/PilotNumbersSection";
import ProductEcosystemSection from "../app/components/ProductEcosystemSection";
import TechStackSection from "../app/components/TechStackSection";
import TeamSection from "../app/components/TeamSection";
import DemoSection from "../app/components/DemoSection";
import FooterSection from "../app/components/FooterSection";
import "../app/globals.css";

export default function App() {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
    >
      <Navbar />
      <main>
        {/* Act 1 — Hook */}
        <HeroSection />

        {/* Act 2 — Problem */}
        <ProblemSection />

        {/* Act 3 — Insight (why InDel is different) */}
        <InsightSection />

        {/* Act 4 — How It Works (zero-touch pipeline) */}
        <HowItWorksSection />

        {/* Act 5A — Signature Features */}
        <FeatureSection />

        {/* Act 5B — Dynamic Pricing Intelligence */}
        <PricingIntelSection />

        {/* Act 6A — Pilot Numbers */}
        <PilotNumbersSection />

        {/* Act 6B — Product Ecosystem (3 dashboards) */}
        <ProductEcosystemSection />

        {/* Act 6C — Architecture */}
        <TechStackSection />

        {/* Act 7A — Demo video */}
        <DemoSection />

        {/* Act 7B — Team + Vision */}
        <TeamSection />
      </main>
      <FooterSection />
    </motion.div>
  );
}
