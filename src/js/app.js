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
    ease: "power1.inOut",
    stagger: 0.5
})

gsap.from("#slideImages", {
    x: 300,
    opacity: 0,
    duration: 2,
    ease: "power1.inOut",
    stagger: 0.5
})