import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateAvailabilityComponent } from './create-availability.component';

describe('CreateAvailabilityComponent', () => {
  let component: CreateAvailabilityComponent;
  let fixture: ComponentFixture<CreateAvailabilityComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateAvailabilityComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CreateAvailabilityComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
