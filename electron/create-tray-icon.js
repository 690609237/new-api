// Create a simple tray icon for macOS
// Run: node create-tray-icon.js

const fs = require('fs');
const { createCanvas } = require('canvas');

function drawModelPassMark(ctx, size) {
  const radius = size * 0.22;
  ctx.beginPath();
  ctx.moveTo(radius, 0);
  ctx.lineTo(size - radius, 0);
  ctx.quadraticCurveTo(size, 0, size, radius);
  ctx.lineTo(size, size - radius);
  ctx.quadraticCurveTo(size, size, size - radius, size);
  ctx.lineTo(radius, size);
  ctx.quadraticCurveTo(0, size, 0, size - radius);
  ctx.lineTo(0, radius);
  ctx.quadraticCurveTo(0, 0, radius, 0);
  ctx.closePath();
  ctx.fillStyle = '#000000';
  ctx.fill();

  ctx.fillStyle = '#ffffff';
  ctx.font = `bold ${Math.round(size * 0.38)}px system-ui`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('MP', size / 2, size / 2);
}

function createTrayIcon() {
  // macOS template icons use a monochrome ModelPass mark and adapt to the menu bar theme.
  const canvas = createCanvas(22, 22);
  const ctx = canvas.getContext('2d');

  ctx.clearRect(0, 0, 22, 22);
  drawModelPassMark(ctx, 22);

  const buffer = canvas.toBuffer('image/png');
  fs.writeFileSync('tray-icon.png', buffer);

  fs.writeFileSync('tray-iconTemplate.png', buffer);
  const retinaCanvas = createCanvas(44, 44);
  drawModelPassMark(retinaCanvas.getContext('2d'), 44);
  fs.writeFileSync('tray-iconTemplate@2x.png', retinaCanvas.toBuffer('image/png'));

  console.log('Tray icon created successfully!');
}

// Check if canvas is installed
try {
  createTrayIcon();
} catch (err) {
  console.log('Canvas module not installed.');
  console.log('For now, creating a placeholder. Install canvas with: npm install canvas');

  // Create a minimal 1x1 transparent PNG as placeholder
  const minimalPNG = Buffer.from([
    0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
    0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
    0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
    0x01, 0x03, 0x00, 0x00, 0x00, 0x25, 0xDB, 0x56,
    0xCA, 0x00, 0x00, 0x00, 0x03, 0x50, 0x4C, 0x54,
    0x45, 0x00, 0x00, 0x00, 0xA7, 0x7A, 0x3D, 0xDA,
    0x00, 0x00, 0x00, 0x01, 0x74, 0x52, 0x4E, 0x53,
    0x00, 0x40, 0xE6, 0xD8, 0x66, 0x00, 0x00, 0x00,
    0x0A, 0x49, 0x44, 0x41, 0x54, 0x08, 0x1D, 0x62,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
    0x00, 0x01, 0x0A, 0x2D, 0xCB, 0x59, 0x00, 0x00,
    0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42,
    0x60, 0x82
  ]);

  fs.writeFileSync('tray-icon.png', minimalPNG);
  console.log('Created placeholder tray icon.');
}
