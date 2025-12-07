import type { Meta, StoryObj } from '@storybook/svelte-vite';
import BrandLogo from './BrandLogo.svelte';
import BrandText from './BrandText.svelte';
import BrandFull from './BrandFull.svelte';

const meta = {
  title: 'Brand/Components',
  component: BrandFull,
  tags: ['autodocs'],
  argTypes: {
    size: {
      control: { type: 'select' },
      options: ['sm', 'md', 'lg', 'xl'],
    },
    layout: {
      control: { type: 'radio' },
      options: ['horizontal', 'vertical'],
    },
  },
} satisfies Meta<BrandFull>;

export default meta;
type Story = StoryObj<typeof meta>;

// === Brand Full Stories ===

export const FullLogoHorizontal: Story = {
  args: {
    size: 'lg',
    layout: 'horizontal',
  },
};

export const FullLogoVertical: Story = {
  args: {
    size: 'lg',
    layout: 'vertical',
  },
};

export const FullLogoSmall: Story = {
  args: {
    size: 'sm',
    layout: 'horizontal',
  },
};

// === Individual Components (via Render render function or specific stories if needed, 
// but Storybook best practice is usually one component per file or using subcomponents. 
// For simplicity in this demo, I'll create separate basic stories for the sub-components 
// by just rendering them in a container if I want, but let's stick to the main export being BrandFull 
// and maybe add separate exports if possible or just rely on BrandFull to show them off.)

// Actually, let's make this file default to BrandFull, but we can have other stories that use the other components.
// Storybook Svelte allows returning specific components in the render function.

export const LogoOnly: Story = {
  render: (args) => ({
    Component: BrandLogo,
    props: {
      size: args.size,
    },
  }),
  args: {
    size: 'lg',
  },
};

export const TextOnly: Story = {
  render: (args) => ({
    Component: BrandText,
    props: {
      size: args.size,
    },
  }),
  args: {
    size: 'lg',
  },
  globals: {
    backgrounds: {
      value: "dark"
    }
  }
};
