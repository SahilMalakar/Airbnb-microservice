// src/infra/database/base.repository.ts
import type { Prisma } from './generated/client.js';

type SoftDeletable = {
  deletedAt?: Date | string | null | Prisma.NullableDateTimeFieldUpdateOperationsInput;
};

type MinimalDelegate<TModel, TWhereUnique, TWhere, TCreateInput, TUpdateInput> = {
  findUnique(args: { where: TWhereUnique }): Promise<TModel | null>;
  findMany(args?: { where?: TWhere }): Promise<TModel[]>;
  create(args: { data: TCreateInput }): Promise<TModel>;
  update(args: { where: TWhereUnique; data: TUpdateInput }): Promise<TModel>;
  delete(args: { where: TWhereUnique }): Promise<TModel>;
};

export class BaseRepository
  <TModel,
  TWhereUnique,
  TWhere,
  TCreateInput,
  TUpdateInput extends SoftDeletable
> {
  constructor(
    protected readonly model: MinimalDelegate<TModel, TWhereUnique, TWhere, TCreateInput, TUpdateInput>
  ) {}

  async findById(where: TWhereUnique) {
    return this.model.findUnique({ where });
  }

  async findAll(where?: TWhere) {
    return where !== undefined
      ? this.model.findMany({ where })
      : this.model.findMany();
  }

  async create(data: TCreateInput) {
    return this.model.create({ data });
  }

  async update(where: TWhereUnique, data: TUpdateInput) {
    return this.model.update({ where, data });
  }

  async softDelete(where: TWhereUnique) {
    return this.model.update({ where, data: { deletedAt: new Date() } as TUpdateInput });
  }

  async restore(where: TWhereUnique) {
    return this.model.update({ where, data: { deletedAt: null } as TUpdateInput });
  }

  async hardDelete(where: TWhereUnique) {
    return this.model.delete({ where });
  }
}