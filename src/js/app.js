import { gsap } from "gsap";
import { SplitText } from "gsap/SplitText";

gsap.registerPlugin(SplitText);

gsap.set("#heading", { opacity: 1 });

let split = SplitText.create("#heading", { type: "chars" });
gsap.from(split.chars, {
    y: 20,
    autoAlpha: 0,
    stagger: 0.05
});

gsap.from("#slideText", {
    x: -300,
    opacity: 0,
    duration: 2,
    ease: "power2.Out",
})

gsap.from("#slideImages", {
    x: 300,
    opacity: 0,
    duration: 2,
    ease: "power2.Out",
})

gsap.from("#fadeIn", {
    opacity: 0,
    duration: 3,
    ease: "power2.Out",
})

gsap.from("#zoom", {
    scale: 0,
    opacity: 0,
    duration: 1.5,
    ease: "power2.Out",
})