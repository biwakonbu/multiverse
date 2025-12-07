import type { Meta, StoryObj } from '@storybook/svelte';
import type { ComponentProps } from 'svelte';
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
} as Meta<typeof BrandFull>;

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

// === Individual Components ===

export const LogoOnly: Story = {
  render: (args: ComponentProps<typeof BrandFull>) => ({
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
  render: (args: ComponentProps<typeof BrandFull>) => ({
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
