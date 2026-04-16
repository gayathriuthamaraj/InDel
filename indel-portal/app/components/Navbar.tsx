import React from "react";

function scrollToSection(e: React.MouseEvent<HTMLAnchorElement, MouseEvent>, href: string) {
  if (href.startsWith("#")) {
    e.preventDefault();
    const id = href.replace("#", "");
    const el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ behavior: "smooth" });
    }
  }
}

export default function Navbar() {
  return (
    <nav className="flex justify-between items-center px-8 py-4 bg-primary-light/60 sticky top-0 z-50 shadow-sm border-b border-primary-light">
      <a href="/" className="font-bold text-primary text-xl tracking-tight">InDel</a>
      <div className="hidden md:flex gap-2">
        <a href="#ecosystem" className="text-primary-dark hover:bg-primary-light hover:text-primary px-3 py-2 rounded transition font-medium" onClick={e => scrollToSection(e, "#ecosystem")}>Product</a>
        <a href="#how-it-works" className="text-primary-dark hover:bg-primary-light hover:text-primary px-3 py-2 rounded transition font-medium" onClick={e => scrollToSection(e, "#how-it-works")}>How it Works</a>
        <a href="#tech" className="text-primary-dark hover:bg-primary-light hover:text-primary px-3 py-2 rounded transition font-medium" onClick={e => scrollToSection(e, "#tech")}>Technology</a>
        <a href="#demo" className="text-primary-dark hover:bg-primary-light hover:text-primary px-3 py-2 rounded transition font-medium" onClick={e => scrollToSection(e, "#demo")}>Demo</a>
        <div className="relative group">
          <button className="text-primary-dark hover:bg-primary-light hover:text-primary px-3 py-2 rounded transition font-medium">More ▾</button>
          <div className="absolute left-0 mt-2 w-40 bg-surface border border-primary-light rounded shadow-lg opacity-0 group-hover:opacity-100 pointer-events-none group-hover:pointer-events-auto transition-opacity duration-200 z-10">
            <a href="#problem" className="block px-4 py-2 text-primary-dark hover:bg-primary-light/60" onClick={e => scrollToSection(e, "#problem")}>Problem</a>
            <a href="#solution" className="block px-4 py-2 text-primary-dark hover:bg-primary-light/60" onClick={e => scrollToSection(e, "#solution")}>Solution</a>
            <a href="#dashboard-preview" className="block px-4 py-2 text-primary-dark hover:bg-primary-light/60" onClick={e => scrollToSection(e, "#dashboard-preview")}>Dashboard</a>
            <a href="#why" className="block px-4 py-2 text-primary-dark hover:bg-primary-light/60" onClick={e => scrollToSection(e, "#why")}>Why</a>
            <a href="#vision" className="block px-4 py-2 text-primary-dark hover:bg-primary-light/60" onClick={e => scrollToSection(e, "#vision")}>Vision</a>
          </div>
        </div>
      </div>
    </nav>
  );
}
